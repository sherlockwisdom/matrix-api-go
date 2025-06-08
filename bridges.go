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
	if bridgeCfg, ok := conf.GetBridgeConfig(b.Name); ok {
		var clientDb = ClientDB{
			username: b.Client.UserID.Localpart(),
			filepath: "db/" + b.Client.UserID.Localpart() + ".db",
		}

		if err := clientDb.Init(); err != nil {
			return err
		}

		var room Rooms
		for _, bridge := range syncingClients.Bridge[b.Client.UserID.Localpart()] {
			if bridge.Name == b.Name {
				room = bridge.Room
				break
			}
		}

		if room.ID == "" {
			return fmt.Errorf("room not found for bridge: %s", b.Name)
		}

		var wg sync.WaitGroup
		if loginCmd, exists := bridgeCfg.Cmd["login"]; exists {
			wg.Add(1)
			go func() {
				since := time.Now().UnixMilli()
				log.Printf("Waiting for events %s %p\n", b.Client.UserID, b.ChEvt)
				for evt := range b.ChEvt {
					if evt.RoomID == b.Room.ID &&
						evt.Sender != b.Client.UserID &&
						evt.Timestamp >= since &&
						evt.Type == event.EventMessage {

						log.Println("Event:", evt)

						failedCmd := bridgeCfg.Cmd["failed"]
						matchesSuccess, err := cfg.CheckSuccessPattern(b.Name, evt.Content.AsMessage().Body)
						if err != nil {
							log.Println("Error checking success pattern:", err)
							b.ChImage <- nil
							break
						}

						if evt.Content.Raw["msgtype"] == "m.notice" {
							if strings.Contains(evt.Content.AsMessage().Body, failedCmd) || matchesSuccess {
								log.Println("Get new notice to failed or success:", evt)
								b.ChImage <- nil
								break
							}
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
				_, err = b.Client.SendText(
					context.Background(),
					id.RoomID(b.Room.ID),
					bridgeCfg.Cmd["cancel"],
				)

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
	username string,
	botName string,
) error {
	clientRooms, err := b.Room.JoinedRooms(b.Client)
	if err != nil {
		return err
	}
	log.Println("Client Rooms: ", clientRooms)

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}
	clientDb.Init()

	managementRoom := false
	var roomId id.RoomID
	for _, clientRoom := range clientRooms {
		managementRoom, err = b.IsManagementRoom(botName, clientRoom)
		if err != nil {
			return err
		}
		roomId = clientRoom

		if managementRoom {
			break
		}

	}
	if !managementRoom {
		roomId, err := b.Room.CreateRoom(
			b.Client, b.Name, botName, RoomTypeManagement, true,
		)
		if err != nil {
			return err
		}
		log.Println("[+] Created room successfully for:", botName, roomId)
	}
	b.Room.ID = roomId
	clientDb.StoreRooms(roomId.String(), b.Name, botName, int(RoomTypeManagement), true)
	log.Println("[+] Stored room successfully for:", botName, roomId)

	return nil
}

func (b *Bridges) IsManagementRoom(bridgeBotName string, roomId id.RoomID) (bool, error) {
	members, err := b.Room.GetRoomMembers(b.Client, roomId)
	if err != nil {
		return false, err
	}

	for _, member := range members {
		if member.String() == bridgeBotName && len(members) == 2 {
			return true, nil
		}
	}

	return false, nil
}
