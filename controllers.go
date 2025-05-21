package main

import (
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

/**

- Create account

- Invite the bridge

- Accept all pending invites

**/

func CreateProcess(
	client *mautrix.Client,
	room *Rooms,
	username string,
	password string,
) error {
	_, err := Create(client, username, password)

	if err != nil {
		return err
	}

	log.Println("[+] Created user: ", username)

	members := "@signalbot:relaysms.me"
	room.Bridge.username = username
	_, err = room.CreateRoom(client, members, roomTypes.Management, true)
	if err != nil {
		return err
	}

	log.Println("[+] Created room successfully")
	client.UserID = id.UserID("@" + username + ":relaysms.me")

	return nil
}

func LoginProcess(
	client *mautrix.Client,
	room *Rooms,
	username string,
	password string,
) error {
	_, err := LoadActiveSessions(client, username, password)
	if err != nil {
		if _, err = Login(client, username, password); err != nil {
			return err
		}
	}
	client.UserID = id.UserID("@" + username + ":relaysms.me")
	room.Bridge.username = username

	return nil
}

// func main() {
// 	password := "M4yHFt$5hW0UuyTv2hdRwtGryHa9$R7z"
// 	homeServer := "https://relaysms.me"

// 	client, err := mautrix.NewClient(homeServer, "", "")
// 	if err != nil {
// 		panic(err)
// 	}

// 	var bridge Bridges
// 	var room = Rooms{
// 		Channel: make(chan *event.Event),
// 		Bridge:  bridge,
// 	}

// 	if len(os.Args) > 1 {
// 		switch os.Args[1] {
// 		case "--create":
// 			username := "sherlock_" + strconv.FormatInt(time.Now().UnixMilli(), 10)
// 			CreateProcess(
// 				client,
// 				&room,
// 				username,
// 				password,
// 			)
// 		case "--login":
// 			username := os.Args[2]
// 			LoginProcess(client, &room, username, password)
// 		default:
// 		}
// 	}

// 	if len(client.AccessToken) < 3 {
// 		log.Fatalf("Client access token expected: > 2, got: %d %v", len(client.AccessToken), client.AccessToken)
// 		return
// 	}

// 	/*
// 		go func() {
// 			var bot Bots = Bots{}
// 			bot.AddDevice(client, roomId, botChannel)
// 		}()
// 	*/

// 	go func() {
// 		room.ListenJoinedRooms(client)
// 	}()

// 	go func() {
// 		err := Sync(client, &room)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// 	<-sigChan
// 	client.StopSync()
// 	fmt.Println("\nShutdown signal received. Exiting...")
// }
