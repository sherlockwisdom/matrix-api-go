package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
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

func (b *Bridges) AddDevice() error {
	conf, err := (&Conf{}).getConf()

	if err != nil {
		return err
	}

	log.Println("Getting configs for:", b.Name)
	if cfg, ok := conf.GetBridgeConfig(b.Name); ok {
		var clientDb = ClientDB{
			username: b.Client.UserID.Localpart(),
			filepath: "db/" + b.Client.UserID.Localpart() + ".db",
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

		var wg sync.WaitGroup
		if loginCmd, exists := cfg.Cmd["login"]; exists {
			wg.Add(1)
			go func() {
				since := time.Now().UnixMilli()
				log.Printf("Waiting for events %s %p\n", b.Client.UserID, b.ChEvt)
				for evt := range b.ChEvt {
					if evt.RoomID == b.Room.ID && evt.Sender != b.Client.UserID && evt.Timestamp >= since &&
						evt.Type == event.EventMessage {
						log.Println("Event:", evt)

						failedCmd := cfg.Cmd["failed"]

						if evt.Content.Raw["msgtype"] == "m.notice" &&
							strings.Contains(evt.Content.AsMessage().Body, failedCmd) {
							log.Println("Get new notice to failed:", evt)
							b.ChImage <- nil
							break
						}

						if event.MessageType.IsMedia(evt.Content.AsMessage().MsgType) {
							url := evt.Content.AsMessage().URL
							file, err := ParseImage(b.Client, string(url))
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
				defer wg.Done()
			}()

			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.Name, loginCmd, b.Room.ID)
			_, err = b.Client.SendText(
				context.Background(),
				id.RoomID(b.Room.ID),
				loginCmd,
			)

			if err != nil {
				return err
			}

		}
		wg.Wait()
	}
	return err
}

func (b *Bridges) JoinRooms(
	client *mautrix.Client,
	username string,
) error {
	clientDB := ClientDB{
		username: username,
		filepath: "db/" + username + ".db",
	}
	clientDB.Init()

	clientRooms, err := b.Room.JoinedRooms(client)
	if err != nil {
		return err
	}
	log.Println("Client Rooms: ", clientRooms)

	for _, entry := range cfg.Bridges {
		for name, config := range entry {
			managementRoom := false
			for _, clientRoom := range clientRooms {
				managementRoom, err = b.IsManagementRoom(config.BotName, clientRoom)
				if err != nil {
					return err
				}

				if managementRoom {
					break
				}

			}
			if !managementRoom {
				roomId, err := b.Room.CreateRoom(
					client, name, config.BotName, RoomTypeManagement, true,
				)
				if err != nil {
					return err
				}
				log.Println("[+] Created room successfully for:", config.BotName, roomId)
			}
		}
	}
	return nil
}

func (b *Bridges) IsManagementRoom(bridgeBotName string, roomId id.RoomID) (bool, error) {
	members, err := b.Room.GetRoomMembers(b.Client, roomId)
	if err != nil {
		return false, err
	}

	for _, member := range members {
		if member.String() == bridgeBotName {
			return true, nil
		}
	}

	return false, nil
}
