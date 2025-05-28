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
	name string
	room Rooms
	ch   chan *event.Event
}

func (b *Bridges) AddDevice(
	client *mautrix.Client,
) ([]byte, error) {
	conf, err := (&Conf{}).getConf()

	if err != nil {
		return []byte{}, err
	}

	if cfg, ok := conf.GetBridgeConfig(b.name); ok {
		log.Println("Getting configs for:", b.name)
		var clientDb = ClientDB{
			username: b.room.User.name,
			filepath: "db/" + b.room.User.name + ".db",
		}

		if err := clientDb.Init(); err != nil {
			return []byte{}, err
		}

		room, err := clientDb.FetchRoomsByMembers(b.name)
		if err != nil {
			return []byte{}, err
		}
		log.Println("Room:", room)

		b.room = room

		go func() {
			if err := Sync(client, b); err != nil {
				log.Println(err)
			}
		}()

		if loginCmd, exists := cfg.Cmd["login"]; exists {
			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.name, loginCmd, b.room.ID)
			_, err = client.SendText(
				context.Background(),
				id.RoomID(b.room.ID),
				loginCmd,
			)

			if err != nil {
				return []byte{}, err
			}

			for evt := range b.ch {
				if evt.Type == event.EventMessage && evt.RoomID == b.room.ID {
					// msg := evt.Content.AsMessage().Body
					url := evt.Content.AsMessage().URL
					file, err := ParseImage(client, string(url))
					if err != nil {
						return []byte{}, nil
					}

					log.Println("New message adding device:", evt.Content.AsMessage().FileName)
					return file, nil
				}
			}

		}
	}

	return []byte{}, err
}
