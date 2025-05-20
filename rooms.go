package main

import (
	"context"
	"fmt"
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
}

func (r *Room) ListenJoinedRooms(
	client *mautrix.Client,
) {
	fmt.Println(">> Begin listening...")
	joinedRooms, err := r.JoinedRooms(client)
	fmt.Printf("[+] Joined rooms: %v\n", joinedRooms)

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
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
	log.Println("Finished listening to rooms...")
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
	fmt.Println("[+] Getting messages for: ", r.id)
	for {
		evt := <-r.channel
		if evt.Type == event.EventMessage {
			fmt.Printf("Room channel parsing %v, %v\n", evt.Sender, evt.RoomID)
			content := evt.Content.Raw

			switch content["msgtype"] {
			case "m.text":
				fmt.Println(">> " + content["body"].(string))
			case "m.image":
				rawImage, err := ParseImage(client, content["url"].(string))
				if err != nil {
					panic(err)
				}

				imageDownloadFilepath := "downloads/rooms/" + content["filename"].(string)
				os.WriteFile(imageDownloadFilepath, rawImage, 0644)
				fmt.Printf("[+] Saved image to room dir: %s\n", imageDownloadFilepath)
			default:
				fmt.Printf("[-] Type not yet implemented: %v\n", content["msgtype"])
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
	fmt.Println("[+] Getting invites for: ", r.id)
	for {
		resp := <-r.channel

		if resp.Content.AsMember().Membership == event.MembershipInvite {
			if resp.StateKey != nil && *resp.StateKey == client.UserID.String() {
				fmt.Printf("Got invite to room: %s from %s\n", resp.RoomID, resp.Sender)
			}
		}
	}
}
