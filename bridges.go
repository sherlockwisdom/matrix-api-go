package main

import (
	"context"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type BridgesInterface interface {
	AddDevice(
		client *mautrix.Client,
		roomId string,
	)
	HandleMessages(*event.Event) (bool, error)
	DefaultRoom() (string, error)
}

type Bridges struct {
	username string
	name     string
	room     Rooms
	ch       chan *event.Event
}

func (b *Bridges) AddDevice(
	client *mautrix.Client,
) (string, error) {
	addDevicePrompt := "!" + b.name + " login"
	log.Printf("[+] %sBridge| Sending message to %v\n", b.name, b.room.ID)

	_, err := client.SendText(
		context.Background(),
		id.RoomID(b.room.ID),
		addDevicePrompt,
	)

	if err != nil {
		panic(err)
	}

	for evt := range b.ch {
		return evt.Content.AsMessage().Body, nil
	}

	return "", nil
}
