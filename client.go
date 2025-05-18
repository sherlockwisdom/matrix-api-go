package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func LoadActiveSessions(client *mautrix.Client) (string, error) {
	fmt.Println("Loading active sessions...")

	dbFilepath := "db/sessions.txt"

	data, err := os.ReadFile(dbFilepath)
	if err != nil {
		return "", err
	}

	accessToken := string(data)
	client.AccessToken = accessToken

	fmt.Printf("Found access token: %v -> %s", data, accessToken)

	return accessToken, nil
}

func Login(client *mautrix.Client, username string, password string) {
	fmt.Printf("Login in as %s\n", username)

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
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Printf("Login successful. Access token: %s\n", resp.AccessToken)
	client.AccessToken = resp.AccessToken

	dbFilepath := "db/sessions.txt"
	err = os.WriteFile(dbFilepath, []byte(client.AccessToken), 0644)

	if err != nil {
		panic(err)
	}
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
	clientDB.Init()
	clientDB.Store([]byte(client.AccessToken))

	err = os.WriteFile(clientDB.filepath, []byte(client.AccessToken), 0644)

	if err != nil {
		return resp.AccessToken, err
	}

	fmt.Printf("User registered successfully. Access token: %s\n", resp.AccessToken)
	return resp.AccessToken, nil
}

func Sync(client *mautrix.Client, botChan chan *event.Event) {
	syncer := mautrix.NewDefaultSyncer()
	client.Syncer = syncer

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		if evt.Type == event.EventMessage {
			// log.Printf(
			// 	"Event: %s | Room: %s | From: %s | Content: %s\n",
			// 	evt.Type, evt.RoomID, evt.Sender, evt.Content.Parsed,
			// )
			botChan <- evt
			return
		}

		// fmt.Printf(
		// 	"Event: %s | Room: %s | From: %s\n",
		// 	evt.Type, evt.RoomID, evt.Sender,
		// )
	})

	go func() {
		if err := client.Sync(); err != nil {
			panic(err)
		}
	}()

}
