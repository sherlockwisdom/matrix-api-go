package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

/**

- Create account

- Invite the bridge

- Accept all pending invites

**/

func CreateProcess(
	client *mautrix.Client,
	room *Room,
	username string,
	password string,
) {
	_, err := Create(client, username, password)

	if err != nil {
		panic(err)
	}

	println("[+] Created user: ", username)

	_, err = room.CreateRoom(client, "@signalbot:relaysms.me")
	if err != nil {
		log.Fatalf("[-] Failed to create room: %v", err)
	}

	println("[+] Created room successfully")
	client.UserID = id.UserID("@" + username + ":relaysms.me")
}

func LoginProcess(
	client *mautrix.Client,
	room *Room,
	username string,
	password string,
) {
	_, err := LoadActiveSessions(client, username)
	if err != nil {
		Login(client, username, password)
	}
	client.UserID = id.UserID("@" + username + ":relaysms.me")
}

func main() {
	password := ".sh@221Bbs"
	homeServer := "https://relaysms.me"

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		panic(err)
	}

	var room Room

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--create":
			username := "sherlock_" + strconv.FormatInt(time.Now().UnixMilli(), 10)
			CreateProcess(
				client,
				&room,
				username,
				password,
			)
		case "--login":
			username := os.Args[2]
			LoginProcess(client, &room, username, password)
		default:
		}
	}

	if len(client.AccessToken) < 3 {
		log.Fatalf("Client access token expected: > 2, got: %d %v", len(client.AccessToken), client.AccessToken)
		return
	}

	botChannel := make(chan *event.Event)
	roomChannel := make(chan *event.Event)

	/*
		go func() {
			var bot Bots = Bots{}
			bot.AddDevice(client, roomId, botChannel)
		}()
	*/

	Sync(client, botChannel)
	room.ListenJoinedRooms(client, roomChannel)

	if err != nil {
		log.Fatalf("Failed to fetched joined rooms %v", err)
	}

	// roomId := "!lqTEAwpbUhXqEsfGzL:relaysms.me"

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutdown signal received. Exiting...")
}
