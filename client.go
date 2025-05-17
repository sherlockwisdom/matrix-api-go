package main

import (
	"context"
	"fmt"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type EventHandlerFunc func(evt *event.Event) error

// Use pointer receiver here:
func (f *EventHandlerFunc) HandleEvent(evt *event.Event) error {
	return (*f)(evt)
}

func main() {
	// Initialize client with homeserver URL
	username := "@sherlock:relaysms.me"
	password := ".sh@221Bbs"
	accessToken := "syt_YWRtaW4_ZWczPEThwbVkUgLGWLAr_4W8jZy"

	client, err := mautrix.NewClient("https://relaysms.me", id.UserID(username), accessToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

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

	syncer := mautrix.NewDefaultSyncer()
	client.Syncer = syncer

	syncer.OnEvent(func(ctx context.Context, evt *event.Event) {
		fmt.Printf("Event: %s | Room: %s | From: %s\n", evt.Type, evt.RoomID, evt.Sender)
	})

	// go func() {
	// 	if err := client.Sync(); err != nil {
	// 		panic(err)
	// 	}
	// }()

	err = client.Sync()
	if err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	// Logout from the session
	_, err = client.Logout(context.Background())
	if err != nil {
		log.Fatalf("Logout failed: %v", err)
	}

	fmt.Println("Logout successful.")
}
