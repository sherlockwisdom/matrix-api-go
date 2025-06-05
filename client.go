package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type SyncingClients struct {
	Bridge   map[string]*Bridges
	Registry map[string]bool
}

type ClientDB struct {
	connection *sql.DB
	username   string
	filepath   string
}

func LoadActiveSessions(
	client *mautrix.Client,
	username string,
	password string,
) (string, error) {
	fmt.Println("Loading active sessions: ", username, password)

	var clientDB ClientDB = ClientDB{
		username: username,
		filepath: "db/" + username + ".db",
	}
	clientDB.Init()
	exists, err := clientDB.Authenticate(username, password)

	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("user does not exist")
	}

	accessToken, err := clientDB.Fetch()
	if err != nil {
		panic(err)
	}

	client.AccessToken = accessToken

	fmt.Printf("Found access token: %v\n", accessToken)

	return accessToken, nil
}

func Login(client *mautrix.Client, username string, password string) (string, error) {
	fmt.Printf("Login in as %s\n", username)

	var clientDB ClientDB = ClientDB{
		username: username,
		filepath: "db/" + username + ".db",
	}
	clientDB.Init()

	identifier := mautrix.UserIdentifier{
		Type: "m.id.user",
		User: username,
	}

	resp, err := client.Login(context.Background(), &mautrix.ReqLogin{
		Type: "m.login.password",
		// User:     id.UserID(username),
		Identifier: identifier,
		Password:   password,
	})
	if err != nil {
		return "", err
	}

	client.AccessToken = resp.AccessToken

	err = clientDB.Store(client.AccessToken, password)

	if err != nil {
		return "", err
	}

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

func Create(client *mautrix.Client, username string, password string) (string, error) {
	fmt.Printf("[+] Creating user: %s\n", username)

	available, err := client.RegisterAvailable(context.Background(), username)
	if err != nil {
		log.Fatalf("Username availability check failed: %v", err)
		return "", err
	}
	if !available.Available {
		log.Fatalf("Username '%s' is already taken", username)
	}

	resp, _, err := client.Register(context.Background(), &mautrix.ReqRegister{
		Username: username,
		Password: password,
		Auth: map[string]interface{}{
			"type": "m.login.dummy",
		},
	})

	if err != nil {
		return resp.AccessToken, err
	}

	var clientDB ClientDB = ClientDB{
		username: username,
		filepath: "db/" + username + ".db",
	}

	client.AccessToken = resp.AccessToken

	clientDB.Init()
	err = clientDB.Store(client.AccessToken, password)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User registered successfully. Access token: %s\n", resp.AccessToken)
	return resp.AccessToken, nil
}

func Sync(
	client *mautrix.Client,
	bridge *Bridges,
) error {
	syncer := mautrix.NewDefaultSyncer()
	client.Syncer = syncer

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		// bridge.ChEvt <- evt
		// bridge.GetInvites(client, evt)
		go func() {
			bridge.ChEvt <- evt
			bridge.GetInvites(client, evt)
		}()
	})

	log.Println("Syncing...")

	if err := client.Sync(); err != nil {
		log.Println("Sync error for user:", err, client.UserID.String())
		return err
	}
	return nil
}

func (b *Bridges) GetInvites(
	client *mautrix.Client,
	evt *event.Event,
) error {
	if evt.Content.AsMember().Membership == event.MembershipInvite {
		log.Println("[+] Getting invites for: ", b.Room.ID)
		if evt.StateKey != nil && *evt.StateKey == client.UserID.String() {
			log.Printf("[+] >> New invite to room: %s from %s\n", evt.RoomID, evt.Sender)
			err := b.Room.Join(client, evt.RoomID)
			if err != nil {
				return err
			}

			if isBridge, err := b.Room.IsBridgeInviteForContact(evt); isBridge {
				log.Println("Bridge message handled -", evt.RoomID)
				log.Println(err)

				var clientDB ClientDB = ClientDB{
					username: b.Room.User.Username,
					filepath: "db/" + b.Room.User.Username + ".db",
				}

				roomName := evt.Content.AsMember().Displayname

				clientDB.Init()
				clientDB.StoreRooms(evt.RoomID.String(), b.Name, roomName, roomTypes.Contact.Parse(), false)
			}
		}
	}
	return nil
}

func SyncAllClients() error {
	for {
		users, err := ks.FetchAllUsers()

		if err != nil {
			return err
		}

		// TODO: make this multi-threaded
		for _, user := range users {
			if syncingClients.Registry[user.Username] {
				continue
			}

			go func(user Users) {
				homeServer := cfg.HomeServer
				client, err := mautrix.NewClient(homeServer, id.NewUserID(user.Username, cfg.HomeServerDomain), user.AccessToken)
				if err != nil {
					log.Println("Error creating bridge for user:", err, user.Username)
					return
				}

				bridge := Bridges{
					ChEvt:   make(chan *event.Event, 1),
					ChImage: make(chan []byte, 1),
					Room: Rooms{
						User: Users{Username: user.Username},
					},
					Client: client,
				}

				syncingClients.Bridge[user.Username] = &bridge
				syncingClients.Registry[user.Username] = true

				err = Sync(client, &bridge)
				if err != nil {
					log.Println("Sync error for user:", err, client.UserID.String())
				}
				syncingClients.Registry[user.Username] = false
			}(user)
		}
		time.Sleep(3 * time.Second)
		log.Printf("Syncing %d clients", len(syncingClients.Bridge))
	}
}
