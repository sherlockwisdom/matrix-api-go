package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
	Name    string
	Room    Rooms
	Client  *mautrix.Client
	ChEvt   chan *event.Event
	ChImage chan []byte
}

func (b *Bridges) AddDevice(
	client *mautrix.Client,
) error {
	conf, err := (&Conf{}).getConf()

	if err != nil {
		return err
	}

	if cfg, ok := conf.GetBridgeConfig(b.Name); ok {
		log.Println("Getting configs for:", b.Name)
		var clientDb = ClientDB{
			username: b.Room.User.name,
			filepath: "db/" + b.Room.User.name + ".db",
		}

		if err := clientDb.Init(); err != nil {
			return err
		}

		room, err := clientDb.FetchRoomsByMembers(b.Name)
		if err != nil {
			return err
		}
		log.Println("Room:", room)

		b.Room = room

		go func() {
			if err := Sync(client, b); err != nil {
				log.Println(err)
			}
		}()

		if loginCmd, exists := cfg.Cmd["login"]; exists {
			go func() {
				since := time.Now().UnixMilli()
				for evt := range b.ChEvt {
					if evt.RoomID == b.Room.ID && evt.Sender != client.UserID && evt.Timestamp >= since &&
						evt.Type == event.EventMessage {
						failedCmd := cfg.Cmd["failed"]

						if evt.Content.Raw["msgtype"] == "m.notice" &&
							strings.Contains(evt.Content.AsMessage().Body, failedCmd) {
							log.Println("Get new notice to failed:", evt)
							b.ChImage <- nil
							break
						}

						if event.MessageType.IsMedia(evt.Content.AsMessage().MsgType) {
							url := evt.Content.AsMessage().URL
							file, err := ParseImage(client, string(url))
							if err != nil {
								fmt.Println(err)
							}

							// return file, nil
							b.ChImage <- file
							log.Println("New message adding device:", evt.Content.AsMessage().FileName)
							continue
						}
					}
				}
			}()

			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.Name, loginCmd, b.Room.ID)
			_, err = client.SendText(
				context.Background(),
				id.RoomID(b.Room.ID),
				loginCmd,
			)

			if err != nil {
				return err
			}

		}

	}

	return err
}
