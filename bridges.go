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
	BotName string
	RoomID  id.RoomID
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

		var wg sync.WaitGroup
		if loginCmd, exists := bridgeCfg.Cmd["login"]; exists {
			wg.Add(1)
			go func() {
				since := time.Now().UnixMilli()
				log.Printf("Waiting for events %s %p\n", b.Client.UserID, b.ChEvt)
				for evt := range b.ChEvt {
					if evt.RoomID == b.RoomID &&
						evt.Sender != b.Client.UserID &&
						evt.Timestamp >= since &&
						evt.Type == event.EventMessage {

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
					b.RoomID,
					bridgeCfg.Cmd["cancel"],
				)

				defer wg.Done()
			}()

			log.Printf("[+] %sBridge| Sending message %s to %v\n", b.Name, loginCmd, b.RoomID)
			_, err = b.Client.SendText(
				context.Background(),
				b.RoomID,
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

func (b *Bridges) JoinRooms() error {
	var managementRoom = false
	if b.RoomID != "" {
		log.Println("Checking management room:", b.RoomID)
		mngRoom, err := b.IsManagementRoom()
		if err != nil {
			return err
		}
		managementRoom = mngRoom
		log.Println("Management room:", managementRoom)
	} else {
		rooms, err := b.Client.JoinedRooms(context.Background())
		if err != nil {
			return err
		}
		log.Println("Joined rooms:", rooms)
		for _, room := range rooms.JoinedRooms {
			tB := &Bridges{
				Client:  b.Client,
				RoomID:  room,
				BotName: b.BotName,
			}
			mngRoom, err := tB.IsManagementRoom()
			if err != nil {
				return err
			}
			if mngRoom {
				b.RoomID = room
				managementRoom = true
				break
			}
		}
	}

	if !managementRoom {
		resp, err := b.Client.CreateRoom(context.Background(), &mautrix.ReqCreateRoom{
			Invite:   []id.UserID{id.UserID(b.BotName)},
			IsDirect: true,
			// Preset:     "private_chat",
			Preset:     "trusted_private_chat",
			Visibility: "private",
		})

		if err != nil {
			return err
		}
		log.Println("[+] Created room successfully for:", b.BotName, resp.RoomID)
		b.RoomID = resp.RoomID
	}

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}
	clientDb.Init()

	clientDb.StoreRooms(b.RoomID.String(), b.Name, b.BotName, true)
	log.Println("[+] Stored room successfully for:", b.BotName, b.RoomID)

	return nil
}

func (b *Bridges) IsManagementRoom() (bool, error) {
	members, err := b.Client.JoinedMembers(context.Background(), b.RoomID)
	log.Println("Members:", members)
	if err != nil {
		return false, err
	}

	if len(members.Joined) == 2 {
		for userID, _ := range members.Joined {
			if userID.String() == b.BotName {
				return true, nil
			}
		}
	}

	return false, nil
}
