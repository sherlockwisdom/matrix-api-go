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
	Name           string
	BotName        string
	RoomID         id.RoomID
	Client         *mautrix.Client
	ChLoginSyncEvt chan *event.Event
	ChImageSyncEvt chan []byte
	ChMsgEvt       chan *event.Event
}

func getBridge(username string, name string) (*Bridges, error) {
	var clientDb = ClientDB{
		username: username,
		filepath: "db/" + username + ".db",
	}

	bridges, err := clientDb.FetchBridgeRooms(name)
	if err != nil {
		return nil, err
	}

	for _, bridge := range bridges {
		if bridge.Name == name {
			bridge.RoomID = bridge.RoomID
			return bridge, nil
		}
	}

	return nil, nil
}

func (b *Bridges) processIncomingLoginMessages(bridgeCfg *BridgeConfig, wg *sync.WaitGroup) {
	since := time.Now().UnixMilli()
	log.Printf("Waiting for events %s %p\n", b.Client.UserID, b.ChLoginSyncEvt)
	for {
		evt := <-b.ChLoginSyncEvt
		if evt.RoomID == b.RoomID &&
			evt.Sender != b.Client.UserID &&
			evt.Timestamp >= since &&
			evt.Type == event.EventMessage {

			failedCmd := bridgeCfg.Cmd["failed"]
			matchesSuccess, err := cfg.CheckSuccessPattern(b.Name, evt.Content.AsMessage().Body)
			if err != nil {
				log.Println("Error checking success pattern:", err)
				b.ChImageSyncEvt <- nil
				break
			}

			if evt.Content.Raw["msgtype"] == "m.notice" {
				if strings.Contains(evt.Content.AsMessage().Body, failedCmd) || matchesSuccess {
					log.Println("Get new notice to failed or success:", evt)
					b.ChImageSyncEvt <- nil
					break
				}
			}

			if event.MessageType.IsMedia(evt.Content.AsMessage().MsgType) {
				url := evt.Content.AsMessage().URL
				file, err := ParseImage(b.Client, string(url))
				if err != nil {
					log.Println("Error parsing image:", err)
					b.ChImageSyncEvt <- nil
					break
				}

				// return file, nil
				b.ChImageSyncEvt <- file
				log.Println("New message adding device:", evt.Content.AsMessage().FileName)
				continue
			}
		}
	}

	defer wg.Done()
}

func (b *Bridges) startNewSession(loginCmd string) error {
	log.Printf("[+] %sBridge| Sending message %s to %v\n", b.Name, loginCmd, b.RoomID)
	_, err := b.Client.SendText(
		context.Background(),
		b.RoomID,
		loginCmd,
	)

	if err != nil {
		return err
	}
	return nil
}

func (b *Bridges) checkActiveSessions() (bool, error) {
	return false, nil
}

func (b *Bridges) AddDevice() error {
	log.Println("Getting configs for:", b.Name)
	bridgeCfg, ok := cfg.GetBridgeConfig(b.Name)
	if !ok {
		return fmt.Errorf("bridge config not found for: %s", b.Name)
	}

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}

	if err := clientDb.Init(); err != nil {
		return err
	}

	loginCmd, exists := bridgeCfg.Cmd["login"]
	if !exists {
		return fmt.Errorf("login command not found for: %s", b.Name)
	}

	bridge, err := getBridge(b.Client.UserID.Localpart(), b.Name)
	if err != nil {
		return err
	}
	b.RoomID = bridge.RoomID

	if bridge.RoomID == "" {
		return fmt.Errorf("room not found for bridge: %s", b.Name)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go b.processIncomingLoginMessages(bridgeCfg, &wg)

	// check if active addding session
	// then send message if no active sessions

	activeSessions, err := b.checkActiveSessions()
	if err != nil {
		return err
	}

	if !activeSessions {
		b.startNewSession(loginCmd)
	}
	wg.Wait()

	return nil
}

func (b *Bridges) JoinRooms() error {
	joinedRooms, err := b.Client.JoinedRooms(context.Background())
	log.Println("Joined rooms:", joinedRooms)

	if err != nil {
		return err
	}

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}
	clientDb.Init()

	for _, room := range joinedRooms.JoinedRooms {
		room := Rooms{
			Client: b.Client,
			ID:     room,
		}

		isManagementRoom, err := room.IsManagementRoom(b.BotName)
		if err != nil {
			return err
		}
		log.Println("Is management room:", room.ID, isManagementRoom)

		if isManagementRoom {
			b.RoomID = room.ID
			break
		}
	}

	if b.RoomID == "" {
		log.Println("[+] Creating management room for:", b.BotName)
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

		b.RoomID = resp.RoomID
	}

	clientDb.StoreRooms(b.RoomID.String(), b.Name, b.BotName, true)
	log.Println("[+] Stored room successfully for:", b.BotName, b.RoomID)

	return nil
}
