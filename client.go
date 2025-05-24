package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

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

	// Login using username and password
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
		log.Fatalf("Logout failed: %v", err)
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
	room *Rooms,
) error {
	syncer := mautrix.NewDefaultSyncer()
	client.Syncer = syncer

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		// log.Println("[+] New message type: ", evt.Type)
		go func() {
			room.Channel <- evt
			room.GetInvites(client, evt)
		}()
	})

	log.Println("Syncing...")
	if err := client.Sync(); err != nil {
		return err
	}

	return nil
}

func (r *Rooms) GetInvites(
	client *mautrix.Client,
	evt *event.Event,
) {
	if evt.Content.AsMember().Membership == event.MembershipInvite {
		log.Println("[+] Getting invites for: ", r.ID)
		if evt.StateKey != nil && *evt.StateKey == client.UserID.String() {
			log.Printf("[+] >> New invite to room: %s from %s\n", evt.RoomID, evt.Sender)
			err := r.Join(client, evt.RoomID)
			if err != nil {
				panic(err)
			}
		}
	}
}
