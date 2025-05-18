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

func JoinedRooms(client *mautrix.Client) ([]id.RoomID, error) {
	resp, err := client.JoinedRooms(context.Background())

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
	}

	return resp.JoinedRooms, err
}

func GetRoomMessages(
	client *mautrix.Client,
	roomId string,
	roomChan chan *event.Event,
) {
	for {
		resp := <-roomChan
		if resp.RoomID == id.RoomID(roomId) {
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
