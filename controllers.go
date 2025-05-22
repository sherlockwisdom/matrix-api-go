package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

func CompleteRun(
	client *mautrix.Client,
	room *Rooms,
) {
	if len(client.AccessToken) < 3 {
		log.Fatalf("Client access token expected: > 2, got: %d %v", len(client.AccessToken), client.AccessToken)
		return
	}

	/*
		go func() {
			var bot Bots = Bots{}
			bot.AddDevice(client, roomId, botChannel)
		}()
	*/

	callback := func(inMd IncomingMessageMetaData, err error) {
		if err != nil {
			log.Println(err)
			return
		}
		switch inMd.Message.Type {
		case "m.text":
			log.Printf(">> %s %v\n", inMd.Type, inMd)
		case "m.image":
			rawImage, err := ParseImage(client, string(inMd.Message.Content.AsMessage().URL))
			if err != nil {
				panic(err)
			}

			filename := inMd.Message.Content.AsMessage().FileName
			if filename == "" {
				filename = inMd.Message.Content.AsMessage().Body
			}
			imageDownloadFilepath := "downloads/rooms/" + filename
			os.WriteFile(imageDownloadFilepath, rawImage, 0644)
			log.Printf("[+] Saved image to room dir: %s\n", imageDownloadFilepath)
		default:
			log.Printf("[-] Type not yet implemented: %v\n", inMd.Message.Content.Raw["msgtype"])
		}
	}
	go func() {
		room.ListenJoinedRooms(client, callback)
	}()

	go func() {
		err := Sync(client, room)
		if err != nil {
			panic(err)
		}
	}()

	messagingRoomId := ""
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			fmt.Printf("[%s]-> ", messagingRoomId)
			text, _ := reader.ReadString('\n')

			if text == "" || text == "\n" {
				continue
			}

			text = strings.TrimSuffix(text, "\n")

			if strings.Contains(text, ">room") {
				st := strings.Split(text, " ")
				messagingRoomId = st[len(st)-1]
				continue
			}

			if messagingRoomId == "" {
				fmt.Println("** Messaging requires a room")
				continue
			}

			room := Rooms{
				ID: id.RoomID(messagingRoomId),
			}

			resp, err := room.SendRoomMessages(client, text)
			if err != nil {
				fmt.Printf("%v\n", err)
			}

			fmt.Println(resp)
		}

	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	client.StopSync()
	fmt.Println("\nShutdown signal received. Exiting...")

	os.Exit(0)
}
