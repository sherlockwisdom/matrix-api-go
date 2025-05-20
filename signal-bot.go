package main

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func ParseImage(client *mautrix.Client, url string) ([]byte, error) {
	fmt.Printf(">>\tParsing image for: %v\n", url)
	contentUrl, err := id.ParseContentURI(url)
	if err != nil {
		panic(err)
	}
	return client.DownloadBytes(context.Background(), contentUrl)
}

func (bot Bots) AddDevice(
	client *mautrix.Client,
	roomId string,
) {
	// client.SendMessageEvent(
	// 	context.Background(),
	// 	id.RoomID(roomId),
	// 	event.EventMessage,
	// 	map[string]interface{}{
	// 		"body":    "Hello world",
	// 		"msgtype": "m.text",
	// 	},
	// )

	fmt.Printf("> Sending message as %v\n", client.UserID)

	_, err := client.SendText(
		context.Background(),
		id.RoomID(roomId),
		"!signal login",
	)

	if err != nil {
		panic(err)
	}

	// for {
	// 	resp := <-botChan
	// 	if resp.RoomID == id.RoomID(roomId) {
	// 		fmt.Printf("Bot channel parsing %v, %v\n", resp.Sender, resp.RoomID)
	// 		content := resp.Content.Raw

	// 		// for key, value := range contents {
	// 		// 	fmt.Printf("\t%s: %v\n", key, value)
	// 		// }
	// 		// fmt.Printf(">> %v -> %s", contents, contents["msgtype"])

	// 		// TODO: figure out how to get the device pairing out back to the users
	// 		if content["msgtype"] == "m.image" {
	// 			rawImage, err := ParseImage(client, content["url"].(string))
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			imageDownloadFilepath := "downloads/" + content["filename"].(string)
	// 			os.WriteFile(imageDownloadFilepath, rawImage, 0644)
	// 			fmt.Printf("[+] Saved image to: %s", imageDownloadFilepath)
	// 		}
	// 	}
	// }

	// fmt.Printf("Expected roomID: %v, got: %v\n", id.RoomID(roomId), resp.RoomID)
	// fmt.Printf("Expected userID: %v, got: %v\n", client.UserID, resp.ToUserID)
}

func (bot Bots) Linked(
	client *mautrix.Client,
	roomId string,
	deviceId string,
) (bool, error) {
	/*
	* `89811738-f1ec-4793-abe5-0f7ea3794685` (+237690663592) - `CONNECTED`
	 */

	fmt.Printf("> Sending message as %v\n", client.UserID)

	_, err := client.SendText(
		context.Background(),
		id.RoomID(roomId),
		"!signal list-logins",
	)

	if err != nil {
		panic(err)
	}

	return true, nil
}
