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

/*
This function adds the user to the database and joins the bridge rooms
*/
func (m *MatrixClient) ProcessActiveSessions(
	password string,
) error {
	userSync := syncingClients.Users[m.Client.UserID.Localpart()]
	if userSync == nil {
		userSync = &UserSync{
			Name:      m.Client.UserID.Localpart(),
			Bridges:   make([]*Bridges, 0),
			Syncing:   false,
			SyncMutex: sync.Mutex{},
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

	bridges, err := clientDB.FetchBridgeRooms(m.Client.UserID.Localpart())
	if err != nil {
		return err
	}

	for _, entry := range cfg.Bridges {
		for name, config := range entry {
			bridge := Bridges{
				Name:    name,
				Client:  m.Client,
				BotName: config.BotName,
			}
			for _, _bridge := range bridges {
				if _bridge.Name == name {
					bridge.RoomID = _bridge.RoomID
				}
			}

			err = bridge.JoinRooms()
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

func (b *Bridges) GetInvites(
	evt *event.Event,
) error {
	if evt.Content.AsMember().Membership == event.MembershipInvite {
		log.Println("[+] Getting invites for: ", b.RoomID)
		if evt.StateKey != nil && *evt.StateKey == b.Client.UserID.String() {
			log.Printf("[+] >> New invite to room: %s from %s\n", evt.RoomID, evt.Sender)
			_, err := b.Client.JoinRoomByID(context.Background(), evt.RoomID)
			if err != nil {
				return err
			}

			room := Rooms{
				Client: b.Client,
				ID:     evt.RoomID,
			}
			if isBridge, err := room.IsBridgeInviteForContact(evt); isBridge {
				log.Println("Bridge message handled -", evt.RoomID)
				log.Println(err)

				var clientDB ClientDB = ClientDB{
					username: b.Client.UserID.Localpart(),
					filepath: "db/" + b.Client.UserID.Localpart() + ".db",
				}

				roomName := *evt.StateKey
				log.Println("roomName:", roomName)

				clientDB.Init()
				clientDB.StoreRooms(evt.RoomID.String(), b.Name, roomName, false)
			}
		}
	}
	return nil
}

func (m *MatrixClient) Sync() error {
	syncer := mautrix.NewDefaultSyncer()
	m.Client.Syncer = syncer

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		bridges := syncingClients.Users[m.Client.UserID.Localpart()].Bridges
		for _, bridge := range bridges {
			bridge.Client = m.Client
			go func() {
				bridge.ChEvt <- evt
			}()

			go func() {
				bridge.GetInvites(evt)
			}()
		}
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

		// TODO: make this multi-threaded
		for _, user := range users {
			userSync := syncingClients.Users[user.Username]
			if userSync == nil {
				userSync = &UserSync{
					Name:      user.Username,
					Bridges:   make([]*Bridges, 0),
					Syncing:   false,
					SyncMutex: sync.Mutex{},
				}
				syncingClients.Users[user.Username] = userSync
			}
			if userSync.Syncing {
				continue
			}
			log.Printf("Syncing %d clients", len(userSync.Bridges)+1)
			wg.Add(1)
			go func(user Users) {
				log.Println("Syncing user:", user.Username, user.AccessToken)
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
					return
				}

				clientDb := ClientDB{
					username: user.Username,
					filepath: "db/" + user.Username + ".db",
				}

				userSync.SyncMutex.Lock()
				clientDb.Init()
				bridges, err := clientDb.FetchBridgeRooms(user.Username)
				userSync.SyncMutex.Unlock()

				log.Println("Bridges:", bridges)
				if len(bridges) == 0 {
					log.Println("No bridges found for user:", user.Username)
					return
				}

				if err != nil {
					log.Println("Error fetching bridge rooms for user:", err, user.Username)
					return
				}

				// for _, bridge := range bridges {
				// 	bridge.Client = client
				// 	bridge.JoinRooms()
				// }
				userSync.Syncing = true
				// syncingClients.Bridge[user.Username] = bridges

				err = mc.Sync()
				if err != nil {
					log.Println("Sync error for user:", err, client.UserID.String())
				}

				userSync.Syncing = false
				delete(syncingClients.Users, user.Username)
				log.Println("Deleted syncing for user:", user.Username)
				wg.Done()
			}(user)
		}
		time.Sleep(3 * time.Second)
	}
}
