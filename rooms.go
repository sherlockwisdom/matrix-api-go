package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Room struct {
	id id.RoomID
}

func (r *Room) JoinedRooms(client *mautrix.Client) ([]id.RoomID, error) {
	resp, err := client.JoinedRooms(context.Background())

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
	}

	return resp.JoinedRooms, err
}

func (r *Room) GetRoomMessages(
	client *mautrix.Client,
	roomChan chan *event.Event,
) {
	for {
		resp := <-roomChan
		if resp.RoomID == id.RoomID(r.id) {
			fmt.Printf("Room channel parsing %v, %v\n", resp.Sender, resp.RoomID)
			content := resp.Content.Raw

			switch content["msgtype"] {
			case "m.text":
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

	r.id = resp.RoomID
	return resp.RoomID, nil
}
