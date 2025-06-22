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
	userSync := syncingClients.Users[m.Client.UserID.Localpart()]
	if userSync == nil {
		userSync = &UserSync{
			Name:       m.Client.UserID.Localpart(),
			MsgBridges: make([]*Bridges, 0),
			Syncing:    false,
			SyncMutex:  sync.Mutex{},
		}
		syncingClients.Users[m.Client.UserID.Localpart()] = userSync
	}

	userSync.SyncMutex.Lock()
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

	defer userSync.SyncMutex.Unlock()
	return nil
}

func (m *MatrixClient) LoadActiveSessionsByAccessToken(accessToken string) (string, error) {
	log.Println("Loading active sessions: ", m.Client.UserID.Localpart(), accessToken)

	userSync := syncingClients.Users[m.Client.UserID.Localpart()]
	if userSync == nil {
		return "", fmt.Errorf("user not found")
	}

	userSync.SyncMutex.Lock()

	var clientDB ClientDB = ClientDB{
		username: m.Client.UserID.Localpart(),
		filepath: "db/" + m.Client.UserID.Localpart() + ".db",
	}
	clientDB.Init()
	exists, err := clientDB.AuthenticateAccessToken(m.Client.UserID.Localpart(), accessToken)
	userSync.SyncMutex.Unlock()

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

func (b *Bridges) HandleLoginEvt(
	evt *event.Event,
) {
	if b.BotName == evt.Sender.String() && (evt.Content.Raw["msgtype"] == "m.notice" ||
		(event.MessageType.IsMedia(evt.Content.AsMessage().MsgType) && evt.Type == event.EventMessage)) {
		b.ChLoginSyncEvt <- evt
	}
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

var syncingUsers = make(map[string][]string)

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

				defer func() {
					delete(syncingClients.Users, user.Username)
					log.Println("Deleted syncing for user:", user.Username)
					wg.Done()
				}()
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

	// insert bridge names into syncingUsers if not already present
	for _, bridge := range bridges {
		if _, ok := syncingUsers[user.Username]; !ok {
			syncingUsers[user.Username] = []string{}
		}
		syncingUsers[user.Username] = append(syncingUsers[user.Username], bridge.Name)
	}

	for _, bridge := range bridges {
		bridge.Client = client
	}

	if len(bridges) == 0 {
		log.Println("No bridges found for user:", user.Username)
		return nil
	}

	if err != nil {
		log.Println("Error fetching bridge rooms for user:", err, user.Username)
		return err
	}

	ch := make(chan *event.Event)
	go func() {
		for {
			evt := <-ch
			log.Println("User:", user.Username, "Received event:", evt)
		}
	}()

	err = mc.Sync(ch)

	if err != nil {
		log.Println("Sync error for user:", err, client.UserID.String())
		return err
	}

	return nil
}

func (m *MatrixClient) syncIncomingMessages(userSync *UserSync) error {
	log.Println("Syncing incoming messages for user:", userSync.Name)
	// bridges := userSync.MsgBridges
	bridges := syncingClients.Users[m.Client.UserID.Localpart()].MsgBridges

	wg := sync.WaitGroup{}
	wg.Add(len(bridges))
	for _, bridge := range bridges {
		log.Println("Listing device for", bridge)
		go func() {
			defer wg.Done()
			devices, err := bridge.ListDevices()
			if err != nil {
				log.Println("Error listing devices:", err)
				return
			}

			log.Println("Devices:", devices)
			go func(bridge1 *Bridges, devices []string) {
				err := m.syncListeners(bridge1)
				if err != nil {
					log.Println("Error syncing listeners:", err)
					return
				}
			}(bridge, devices)
		}()
	}
	log.Println("Done listing....")
	wg.Wait()
	return nil
}

func (m *MatrixClient) syncStoreCreateRooms(userSync *UserSync, evt *event.Event, bridge *Bridges) error {
	incomingMessage := IncomingMessage{
		RoomID: evt.RoomID,
		Sender: evt.Sender,
	}
	clientDb := ClientDB{
		username: bridge.Client.UserID.Localpart(),
		filepath: "db/" + bridge.Client.UserID.Localpart() + ".db",
	}
	if userSync == nil {
		userSync = &UserSync{
			Name:       bridge.Client.UserID.Localpart(),
			MsgBridges: make([]*Bridges, 0),
		}
	}
	userSync.SyncMutex.Lock()
	clientDb.Init()

	err := clientDb.StoreRooms(
		incomingMessage.RoomID.String(),
		bridge.Name,
		incomingMessage.Sender.String(),
		false,
	)
	if err != nil {
		log.Println("Error storing room:", err)
	}
	userSync.SyncMutex.Unlock()
	return nil
}

func (m *MatrixClient) syncListeners(bridge *Bridges) error {
	devices, err := bridge.ListDevices()
	if err != nil {
		log.Println("Error listing devices:", err)
		return err
	}

	for {
		evt := <-bridge.ChMsgEvt
		if evt.RoomID == "" {
			continue
		}

		matchBridge, err := cfg.CheckUsernameTemplate(bridge.Name, evt.Sender.String())
		if err != nil {
			log.Println("Error checking if bridge is bot:", err)
		}

		if !matchBridge {
			continue
		}

		for _, device := range devices {
			deviceName, err := sanitizeContact(device)
			if err != nil {
				log.Println("Failed to sanitize contact", deviceName)
			}
			formattedUsername, err := cfg.FormatUsername(bridge.Name, deviceName)
			log.Println("Formatted username:", formattedUsername, evt.Sender.String())
			if err != nil {
				log.Println("Error formatting username:", err)
				continue
			}
			if formattedUsername == evt.Sender.String() {
				continue
			}
		}

		if devices == nil {
			log.Println("No devices found for bridge:", bridge.Name)
			continue
		}

		go func(bridge2 *Bridges) {
			room := Rooms{
				Client: bridge2.Client,
				ID:     evt.RoomID,
			}
			isManagementRoom, err := room.IsManagementRoom(bridge2.BotName)
			if err != nil {
				log.Println("Error checking if bridge is management room:", err)
			}

			isBridgeBot := func() bool {
				for _, device := range devices {
					if device == evt.Sender.String() {
						return true
					}
				}
				return false
			}()
			isClientUser := evt.Sender.String() == m.Client.UserID.String()

			if evt.Type == event.EventMessage && !isManagementRoom && !isBridgeBot && !isClientUser {
				// log.Println("Storing room:", evt.RoomID, evt.Sender.String(), bridge2.Client.UserID.String())
				if evt.Sender.String() != bridge2.Client.UserID.String() {
					userSync := syncingClients.Users[bridge2.Client.UserID.Localpart()]
					err := m.syncStoreCreateRooms(userSync, evt, bridge2)
					if err != nil {
						log.Println("Error storing room:", err)
					}
				}
			}
		}(bridge)
	}

}
