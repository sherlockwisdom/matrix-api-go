package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func main() {
	// Initialize client with homeserver URL
	username := "@sherlock:relaysms.me"
	password := ".sh@221Bbs"
	accessToken := "syt_YWRtaW4_ZWczPEThwbVkUgLGWLAr_4W8jZy"

	client, err := mautrix.NewClient("https://relaysms.me", id.UserID(username), accessToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	botChannel := make(chan *event.Event)

	_, err = LoadActiveSessions(client)
	if err != nil {
		Login(client, username, password)
	}

	Sync(client, botChannel)

	resp, err := JoinedRooms(client)

	for index, a := range resp {
		fmt.Println(index, a)
	}

	roomId := "!lqTEAwpbUhXqEsfGzL:relaysms.me"

	go func() {
		var bot Bots = Bots{}
		bot.AddDevice(client, roomId, botChannel)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutdown signal received. Exiting...")
}
