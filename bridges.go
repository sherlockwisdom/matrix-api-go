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
	conf, err := (&Conf{}).getConf()

	if err != nil {
		return "", err
	}

	if cfg, ok := conf.GetBridgeConfig(b.name); ok {
		var clientDb = ClientDB{
			username: b.room.User.name,
			filepath: "db/" + b.room.User.name + ".db",
		}

		if err := clientDb.Init(); err != nil {
			return "", err
		}

		room, err := clientDb.FetchRoomsByMembers(b.name)
		if err != nil {
			return "", err
		}

		b.room = room
		if loginCmd, exists := cfg.Cmd["login"]; exists {
			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.name, loginCmd, b.room.ID)
			_, err = client.SendText(
				context.Background(),
				id.RoomID(b.room.ID),
				loginCmd,
			)

			if err != nil {
				return "", err
			}

			for evt := range b.ch {
				return evt.Content.AsMessage().Body, nil
			}

		}
	}

	return "", nil
}
