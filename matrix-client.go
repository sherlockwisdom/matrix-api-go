package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type SyncingClients struct {
	Users map[string]*UserSync
}

type ClientDB struct {
	connection *sql.DB
	username   string
	filepath   string
}

type MatrixClient struct {
	Client *mautrix.Client
}

type IncomingMessage struct {
	RoomID  id.RoomID
	Sender  id.UserID
	Content event.Content
}

/*
This function adds the user to the database and joins the bridge rooms
*/
func (m *MatrixClient) ProcessActiveSessions(
	password string,
) error {
	log.Println("Processing active sessions for user:", m.Client.UserID.Localpart())
	var clientDB ClientDB = ClientDB{
		username: m.Client.UserID.Localpart(),
		filepath: "db/" + m.Client.UserID.Localpart() + ".db",
	}
	clientDB.Init()

	if m.Client.AccessToken != "" && m.Client.UserID != "" && password != "" {
		err := ks.CreateUser(m.Client.UserID.Localpart(), m.Client.AccessToken)
		if err != nil {
			return err
		}

		err = clientDB.Store(m.Client.AccessToken, password)

		if err != nil {
			return err
		}
	}

	for _, entry := range cfg.Bridges {
		for name, config := range entry {
			bridge := Bridges{
				Name:    name,
				Client:  m.Client,
				BotName: config.BotName,
			}

			err := bridge.JoinRooms()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MatrixClient) LoadActiveSessionsByAccessToken(accessToken string) (string, error) {
	log.Println("Loading active sessions: ", m.Client.UserID.Localpart(), accessToken)

	var clientDB ClientDB = ClientDB{
		username: m.Client.UserID.Localpart(),
		filepath: "db/" + m.Client.UserID.Localpart() + ".db",
	}
	clientDB.Init()
	exists, err := clientDB.AuthenticateAccessToken(m.Client.UserID.Localpart(), accessToken)

	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("access token does not exist")
	}

	return accessToken, nil
}

func (m *MatrixClient) LoadActiveSessions(
	password string,
) (string, error) {
	log.Println("Loading active sessions: ", m.Client.UserID.Localpart(), password)

	var clientDB ClientDB = ClientDB{
		username: m.Client.UserID.Localpart(),
		filepath: "db/" + m.Client.UserID.Localpart() + ".db",
	}
	clientDB.Init()
	exists, err := clientDB.Authenticate(m.Client.UserID.Localpart(), password)

	if err != nil {
		return "", err
	}

	if !exists {
		return "", nil
	}

	return clientDB.Fetch()
}

func (m *MatrixClient) Login(password string) (string, error) {
	log.Printf("Login in as %s\n", m.Client.UserID.String())

	identifier := mautrix.UserIdentifier{
		Type: "m.id.user",
		User: m.Client.UserID.String(),
	}

	resp, err := m.Client.Login(context.Background(), &mautrix.ReqLogin{
		Type:       "m.login.password",
		Identifier: identifier,
		Password:   password,
	})
	if err != nil {
		return "", err
	}
	m.Client.AccessToken = resp.AccessToken

	return resp.AccessToken, nil
}

func Logout(client *mautrix.Client) error {
	// Logout from the session
	_, err := client.Logout(context.Background())
	if err != nil {
		log.Printf("Logout failed: %v\n", err)
	}

	// TODO: delete the session file

	fmt.Println("Logout successful.")
	return err
}

func (m *MatrixClient) Create(username string, password string) (string, error) {
	fmt.Printf("[+] Creating user: %s\n", username)

	_, err := m.Client.RegisterAvailable(context.Background(), username)
	if err != nil {
		return "", err
	}
	// if !available.Available {
	// 	log.Fatalf("Username '%s' is already taken", username)
	// }

	resp, _, err := m.Client.Register(context.Background(), &mautrix.ReqRegister{
		Username: username,
		Password: password,
		Auth: map[string]interface{}{
			"type": "m.login.dummy",
		},
	})

	if err != nil {
		return "", err
	}

	return resp.AccessToken, nil
}

func (b *Bridges) GetInvites(
	evt *event.Event,
) error {
	if evt.Content.AsMember().Membership == event.MembershipInvite {
		if evt.StateKey != nil && *evt.StateKey == b.Client.UserID.String() {
			_, err := b.Client.JoinRoomByID(context.Background(), evt.RoomID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MatrixClient) Sync(ch chan *event.Event) error {
	syncer := mautrix.NewDefaultSyncer()
	m.Client.Syncer = syncer

	// syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		ch <- evt
	})

	if err := m.Client.Sync(); err != nil {
		return err
	}
	return nil
}

func (m *MatrixClient) SyncAllClients() error {
	log.Println("Syncing all clients")
	var wg sync.WaitGroup

	for {
		users, err := ks.FetchAllUsers()

		if err != nil {
			return err
		}

		for _, user := range users {
			if _, ok := syncingUsers[user.Username]; ok && len(syncingUsers[user.Username]) > 0 {
				continue
			} else {
				syncingUsers[user.Username] = []string{}
			}

			wg.Add(1)

			go func(user Users) {
				err := m.syncClient(user) //blocking
				if err != nil {
					log.Println("Error syncing client:", err)
					return
				}

			}(user)
		}

		time.Sleep(3 * time.Second)
	}
}

func (m *MatrixClient) syncClient(user Users) error {
	homeServer := cfg.HomeServer
	client, err := mautrix.NewClient(
		homeServer,
		id.NewUserID(user.Username, cfg.HomeServerDomain),
		user.AccessToken,
	)
	mc := MatrixClient{
		Client: client,
	}
	if err != nil {
		log.Println("Error creating bridge for user:", err, user.Username)
		return err
	}

	clientDb := ClientDB{
		username: user.Username,
		filepath: "db/" + user.Username + ".db",
	}

	clientDb.Init()
	bridges, err := clientDb.FetchBridgeRooms(user.Username)

	ch := make(chan *event.Event)
	go func() {
		for {
			evt := <-ch
			go m.processIncomingEvents(evt)
		}
	}()

	// insert bridge names into syncingUsers if not already present
	go func() {
		for _, bridge := range bridges {
			bridge.Client = client
			bridge.Client.StateStore = mautrix.NewMemoryStateStore()
			if _, ok := syncingUsers[user.Username]; !ok {
				syncingUsers[user.Username] = []string{}
			}
			syncingUsers[user.Username] = append(syncingUsers[user.Username], bridge.Name)
			if _, ok := ClientDevices[user.Username]; !ok {
				ClientDevices[user.Username] = make(map[string][]string)
			} else if _, ok := ClientDevices[user.Username][bridge.Name]; !ok {
				ClientDevices[user.Username][bridge.Name] = make([]string, 0)
			}

			devices, err := bridge.ListDevices()
			log.Println("Devices for bridge:", bridge.Name, devices)

			if err != nil {
				log.Println("Error listing devices for user:", err, user.Username)
				continue
			}
			ClientDevices[user.Username][bridge.Name] = devices
		}

		if len(bridges) == 0 {
			log.Println("No bridges found for user:", user.Username)
			return
		}
		for _, bridge := range bridges {
			go func(bridge *Bridges) {
				bridge.CreateContactRooms()
				log.Println("Joined member rooms for bridge:", bridge.Name)
				// wg.Done()
			}(bridge)
		}
	}()

	err = mc.Sync(ch)

	if err != nil {
		log.Println("Sync error for user:", err, client.UserID.String())
		return err
	}

	return nil
}

func (m *MatrixClient) processIncomingEvents(evt *event.Event) error {
	for _, subscriber := range EventSubscribers {
		if len(subscriber.ExcludeMsgTypes) > 0 {
			for _, excludeMsgType := range subscriber.ExcludeMsgTypes {
				if excludeMsgType == evt.Content.AsMessage().MsgType {
					continue
				}
			}
		}

		if subscriber.MsgType == nil || *subscriber.MsgType == evt.Content.AsMessage().MsgType {
			if subscriber.RoomID != "" && subscriber.RoomID != evt.RoomID {
				continue
			}

			if subscriber.Since != nil && subscriber.Since.After(time.Unix(evt.Timestamp, 0)) {
				continue
			}

			subscriber.Callback(evt)
		}

	}

	return nil
}
