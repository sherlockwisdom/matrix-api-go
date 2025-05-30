package main

import (
	"context"
	"fmt"
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
	name    string
	room    Rooms
	chEvt   chan *event.Event
	chImage chan []byte
}

func (b *Bridges) AddDevice(
	client *mautrix.Client,
) error {
	conf, err := (&Conf{}).getConf()

	if err != nil {
		return err
	}

	if cfg, ok := conf.GetBridgeConfig(b.name); ok {
		log.Println("Getting configs for:", b.name)
		var clientDb = ClientDB{
			username: b.room.User.name,
			filepath: "db/" + b.room.User.name + ".db",
		}

		if err := clientDb.Init(); err != nil {
			return err
		}

		room, err := clientDb.FetchRoomsByMembers(b.name)
		if err != nil {
			return err
		}
		log.Println("Room:", room)

		b.room = room

		go func() {
			if err := Sync(client, b); err != nil {
				log.Println(err)
			}
		}()

		if loginCmd, exists := cfg.Cmd["login"]; exists {
			go func() {
				for evt := range b.chEvt {
					if evt.Type == event.EventMessage && evt.RoomID == b.room.ID && evt.Sender != client.UserID {
						// msg := evt.Content.AsMessage().Body
						fmt.Println(evt.Content)
						if event.MessageType.IsMedia(evt.Content.AsMessage().MsgType) {
							url := evt.Content.AsMessage().URL
							file, err := ParseImage(client, string(url))
							if err != nil {
								fmt.Println(err)
							}

							log.Println("New message adding device:", evt.Content.AsMessage().FileName)
							// return file, nil
							b.chImage <- file
						}
					}
				}

			}()

			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.name, loginCmd, b.room.ID)
			_, err = client.SendText(
				context.Background(),
				id.RoomID(b.room.ID),
				loginCmd,
			)

			if err != nil {
				return err
			}

		}

	}

	return err
}
