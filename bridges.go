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

type Bridges struct {
	Name    string
	BotName string
	RoomID  id.RoomID
	Client  *mautrix.Client
}

func (b *Bridges) processIncomingLoginMessages(bridgeCfg *BridgeConfig, ch chan []byte) {
	since := time.Now().UnixMilli() - (100 * 1000)

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}

	if err := clientDb.Init(); err != nil {
		log.Println("Error initializing client db:", err)
		return
	}

	eventSubscriber := EventSubscriber{}
	for _, subscriber := range EventSubscribers {
		if subscriber.Name ==
			ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain) &&
			subscriber.MsgType == nil {
			eventSubscriber = subscriber
		}
	}

	if eventSubscriber.Name == "" {
		eventSubscriber = EventSubscriber{
			Name:    ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain),
			MsgType: nil,
			Callback: func(evt *event.Event) {
				log.Println("Received bridge event:", evt.Content.AsMessage().Body, evt.RoomID, b.RoomID, evt.Sender, " -> ", b.Client.UserID)
				if evt.RoomID == b.RoomID &&
					evt.Sender != b.Client.UserID &&
					evt.Timestamp >= since &&
					evt.Type == event.EventMessage {

					failedCmd := bridgeCfg.Cmd["failed"]

					matchesSuccess, err := cfg.CheckSuccessPattern(b.Name, evt.Content.AsMessage().Body)

					if err != nil {
						log.Println("Error checking success pattern:", err)
						ch <- nil
					}

					if evt.Content.Raw["msgtype"] == "m.notice" {
						if strings.Contains(evt.Content.AsMessage().Body, failedCmd) || matchesSuccess {
							log.Println("Get new notice to failed or success:", evt)
							ch <- nil
						}
					}

					if evt.Content.AsMessage().MsgType.IsMedia() {
						url := evt.Content.AsMessage().URL
						file, err := ParseImage(b.Client, string(url))
						if err != nil {
							log.Println("Error parsing image:", err)
							ch <- nil
						}

						// return file, nil
						clientDb.StoreActiveSessions(b.Client.UserID.Localpart(), file)
						ch <- file
						log.Println("New message adding device:", evt.Content.AsMessage().FileName)
					}
				}
			},
		}
		EventSubscribers = append(EventSubscribers, eventSubscriber)
	}
}

func (b *Bridges) startNewSession(cmd string) error {
	log.Printf("[+] %sBridge| Sending message %s to %v\n", b.Name, cmd, b.RoomID)
	_, err := b.Client.SendText(
		context.Background(),
		b.RoomID,
		cmd,
	)

	if err != nil {
		log.Println("Error sending message:", err)
		return err
	}
	return nil
}

func (b *Bridges) checkActiveSessions() (bool, error) {
	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}

	if err := clientDb.Init(); err != nil {
		return false, err
	}

	if IsActiveSessionsExpired(&clientDb, b.Client.UserID.Localpart()) {
		return false, nil
	}

	return true, nil
}

func (b *Bridges) AddDevice(ch chan []byte) error {
	log.Println("Getting configs for:", b.Name, b.RoomID)
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

	go b.processIncomingLoginMessages(bridgeCfg, ch)

	activeSessions, err := b.checkActiveSessions()
	if err != nil {
		log.Println("Failed checking active sessions", err)
		return err
	}

	if !activeSessions {
		log.Println("No active sessions found, removing active sessions")
		clientDb.RemoveActiveSessions(b.Client.UserID.Localpart())
		err := b.startNewSession(loginCmd)
		if err != nil {
			log.Println("Failed starting new session", err)
			return err
		}
	}

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

func (b *Bridges) ListDevices(ch chan []string) ([]string, error) {
	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain)
	eventType := event.MsgNotice
	eventSubscriber := EventSubscriber{
		Name:    eventSubName,
		MsgType: &eventType,
		Callback: func(event *event.Event) {
			log.Println("Received event:", event.Content.AsMessage().Body, event.RoomID, b.RoomID, event.Sender, " -> ", b.Client.UserID)

			ch <- strings.Split(event.Content.AsMessage().Body, "\n")

			defer func() {
				for index, subscriber := range EventSubscribers {
					if subscriber.Name == eventSubName {
						EventSubscribers = append(EventSubscribers[:index], EventSubscribers[index+1:]...)
						log.Println("Removed event subscriber:", eventSubName)
						break
					}
				}
			}()
		},
	}

	EventSubscribers = append(EventSubscribers, eventSubscriber)

	bridgeCfg, ok := cfg.GetBridgeConfig(b.Name)
	if !ok {
		return nil, fmt.Errorf("bridge config not found for: %s", b.Name)
	}
	log.Println("Event subscriber name:", eventSubName)

	_, err := b.Client.SendText(
		context.Background(),
		b.RoomID,
		bridgeCfg.Cmd["devices"],
	)

	if err != nil {
		return nil, err
	}

	devices := <-ch

	return devices, nil
}
