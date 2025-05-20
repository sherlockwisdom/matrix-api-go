package main

import (
	"context"
	"log"
	"os"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Room struct {
	id      id.RoomID
	channel chan *event.Event
	bot     Bots
}

func (r *Room) ListenJoinedRooms(
	client *mautrix.Client,
) {
	log.Println(">> Begin listening...")
	joinedRooms, err := r.JoinedRooms(client)
	log.Printf("[+] Joined rooms: %v\n", joinedRooms)

	if err != nil {
		log.Fatalf("[-] Failed to fetch rooms: %v", err)
	}

	var wg sync.WaitGroup
	for _, roomId := range joinedRooms {
		wg.Add(1)
		var room = Room{
			id:      id.RoomID(roomId),
			channel: r.channel,
		}
		go func() {
			room.GetRoomMessages(client)
			defer wg.Done()
		}()
		go func() {
			room.GetInvites(client)
			defer wg.Done()
		}()
	}
	wg.Wait()
	log.Println("[-] Finished listening to rooms...")
}

func (r *Room) JoinedRooms(
	client *mautrix.Client,
) ([]id.RoomID, error) {
	resp, err := client.JoinedRooms(context.Background())

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
	}

	return resp.JoinedRooms, err
}

func (r *Room) GetRoomMessages(
	client *mautrix.Client,
) {
	log.Println("[+] Getting messages for: ", r.id)
	for {
		evt := <-r.channel
		if evt.Type == event.EventMessage {
			log.Printf("[*] Room channel parsing %v, %v\n", evt.Sender, evt.RoomID)
			content := evt.Content.Raw

			if isBot, _ := r.bot.HandleMessage(evt); isBot {
				return
			}

			switch content["msgtype"] {
			case "m.text":
				log.Println("[+] MSG:", content["body"].(string))
			case "m.image":
				log.Println("[+] saving image", evt.Content.Raw)
				rawImage, err := ParseImage(client, content["url"].(string))
				if err != nil {
					panic(err)
				}

				filename := content["filename"]
				if filename == nil {
					filename = content["body"]
				}
				imageDownloadFilepath := "downloads/rooms/" + filename.(string)
				os.WriteFile(imageDownloadFilepath, rawImage, 0644)
				log.Printf("[+] Saved image to room dir: %s\n", imageDownloadFilepath)
			default:
				log.Printf("[-] Type not yet implemented: %v\n", content["msgtype"])
			}
		}
	}
}

func (r *Room) CreateRoom(
	client *mautrix.Client,
	recipient string,
) (id.RoomID, error) {
	resp, err := client.CreateRoom(context.Background(), &mautrix.ReqCreateRoom{
		Invite:   []id.UserID{id.UserID(recipient)},
		IsDirect: true,
		// Preset:     "private_chat",
		Preset:     "trusted_private_chat",
		Visibility: "private",
	})

	if err != nil {
		return "", err
	}

	return resp.RoomID, nil
}

func (r *Room) GetInvites(
	client *mautrix.Client,
) {
	log.Println("[+] Getting invites for: ", r.id)
	for {
		evt := <-r.channel

		if evt.Content.AsMember().Membership == event.MembershipInvite {
			if evt.StateKey != nil && *evt.StateKey == client.UserID.String() {
				log.Printf("[+] >> New invite to room: %s from %s\n", evt.RoomID, evt.Sender)
				err := r.Join(client, evt.RoomID)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (r *Room) Join(
	client *mautrix.Client,
	roomId id.RoomID,
) error {
	log.Println("[*] Joining room:", roomId)
	_, err := client.JoinRoomByID(context.Background(), roomId)
	return err
}
