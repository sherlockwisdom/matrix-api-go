package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
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

func (bot *Bots) HandleMessage(evt *event.Event) (bool, error) {
	// check room
	// check template

	if evt.Type == event.EventMessage {
		if strings.Contains(evt.Sender.String(), "@signal_") {
			log.Println("[+] BOT| New message:", evt.Content.AsMessage().Body)

			return true, nil
		}
	}
	return false, nil
}

func (bot *Bots) AddDevice(
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
