package main

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (bot Bots) AddDevice(
	client *mautrix.Client,
	roomId string,
	botChan chan *event.Event,
) error {
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

	resp := <-botChan

	for _resp := range botChan {
		if _resp.RoomID == id.RoomID(roomId) {
			fmt.Printf("Bot channel parsing %v, %v\n", _resp.Sender, _resp.RoomID)
			contents := resp.Content.Raw

			for key, value := range contents {
				fmt.Printf("\t%s: %v\n", key, value)
			}

			return nil
		}
	}

	// fmt.Printf("Expected roomID: %v, got: %v\n", id.RoomID(roomId), resp.RoomID)
	// fmt.Printf("Expected userID: %v, got: %v\n", client.UserID, resp.ToUserID)

	return err
}
