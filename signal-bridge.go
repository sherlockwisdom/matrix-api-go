package main

import (
	"context"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func (bridge *Bridges) AddDevice(
	client *mautrix.Client,
	roomId string,
) {
	log.Printf("[+] BOT| Sending message as %v\n", client.UserID)

	_, err := client.SendText(
		context.Background(),
		id.RoomID(roomId),
		"!signal login",
	)

	if err != nil {
		panic(err)
	}
}
