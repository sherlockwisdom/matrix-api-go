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
	Name       string
	BotName    string
	DeviceName string
	RoomID     id.RoomID
	Client     *mautrix.Client
}

func (b *Bridges) ProcessIncomingLoginDaemon(bridgeCfg *BridgeConfig) {
	log.Println("Processing incoming login daemon for:", b.Name)
	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}

	if err := clientDb.Init(); err != nil {
		log.Println("Error initializing client db:", err)
		return
	}

	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain) + "+loginDaemon"
	eventSubscriber := EventSubscriber{
		Name:    eventSubName,
		MsgType: nil,
		ExcludeMsgTypes: []event.MessageType{
			event.MsgText,
		},
		RoomID: b.RoomID,
		Callback: func(evt *event.Event) {
			log.Println("Received event in login:", evt.RoomID, evt.Sender, evt.Timestamp, evt.Type)
			if evt.Sender != b.Client.UserID && evt.Type == event.EventMessage {
				failedCmd := bridgeCfg.Cmd["failed"]

				matchesSuccess, err := cfg.CheckSuccessPattern(b.Name, evt.Content.AsMessage().Body)

				if err != nil {
					clientDb.RemoveActiveSessions(b.Client.UserID.Localpart())
				}

				if evt.Content.Raw["msgtype"] == "m.notice" {
					if strings.Contains(evt.Content.AsMessage().Body, failedCmd) || matchesSuccess {
						clientDb.RemoveActiveSessions(b.Client.UserID.Localpart())
					}
				}

				if evt.Content.AsMessage().MsgType.IsMedia() {
					url := evt.Content.AsMessage().URL
					file, err := ParseImage(b.Client, string(url))
					if err != nil {
						log.Println("Error parsing image:", err)
						clientDb.RemoveActiveSessions(b.Client.UserID.Localpart())
					}

					// return file, nil
					clientDb.StoreActiveSessions(b.Client.UserID.Localpart(), file)
				}
			}

			// defer func() {
			// 	for index, subscriber := range EventSubscribers {
			// 		if subscriber.Name == eventSubName {
			// 			EventSubscribers = append(EventSubscribers[:index], EventSubscribers[index+1:]...)
			// 			break
			// 		}
			// 	}
			// }()
		},
	}
	EventSubscribers = append(EventSubscribers, eventSubscriber)
}

func (b *Bridges) processIncomingLoginMessages(ch *chan []byte) {
	since := time.Now().UTC().Add(-2 * time.Minute)

	var clientDb = ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}

	if err := clientDb.Init(); err != nil {
		log.Println("Error initializing client db:", err)
		return
	}

	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain) + "+login"
	eventSubscriber := EventSubscriber{}
	for _, subscriber := range EventSubscribers {
		if subscriber.Name == eventSubName {
			// eventSubscriber = subscriber
			log.Println("Event subscriber already exists for:", eventSubName)
			return
		}
	}

	noticeType := event.MsgNotice
	eventSubscriber = EventSubscriber{
		Name:    eventSubName,
		MsgType: &noticeType,
		Since:   &since,
		RoomID:  b.RoomID,
		Callback: func(evt *event.Event) {
			log.Println("New notice for login", evt.RoomID, evt.Sender, evt.Timestamp, evt.Type)
			if evt.Sender != b.Client.UserID && evt.Type == event.EventMessage {
				matchesOngoing, err := cfg.CheckOngoingPattern(b.Name, evt.Content.AsMessage().Body)

				if err != nil {
					log.Println("Error checking ongoing pattern:", err)
					*ch <- nil
				}

				if matchesOngoing {
					time.Sleep(3 * time.Second)
					sessions, _, err := clientDb.FetchActiveSessions(b.Client.UserID.Localpart())
					if err != nil {
						log.Println("Error fetching ongoing sessions:", err)
						*ch <- nil
					}

					*ch <- sessions
				}
			}

			// defer func() {
			// 	for index, subscriber := range EventSubscribers {
			// 		if subscriber.Name == eventSubName {
			// 			EventSubscribers = append(EventSubscribers[:index], EventSubscribers[index+1:]...)
			// 			break
			// 		}
			// 	}
			// }()
		},
	}
	EventSubscribers = append(EventSubscribers, eventSubscriber)
	log.Println("Added event subscriber for:", eventSubscriber)
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

	activeSessions, _, err := clientDb.FetchActiveSessions(b.Client.UserID.Localpart())
	if err != nil {
		return false, err
	}

	if len(activeSessions) == 0 {
		return false, nil
	}

	return true, nil
}

func (b *Bridges) AddDevice(ch *chan []byte) error {
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

	b.processIncomingLoginMessages(ch)
	log.Println("Processed incoming login messages for:", b.Name)

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

func (b *Bridges) JoinManagementRooms() error {
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

	clientDb.StoreRooms(b.RoomID.String(), b.Name, "", b.BotName, true)
	log.Println("[+] Stored room successfully for:", b.BotName, b.RoomID)

	return nil
}

func (b *Bridges) ListDevices() ([]string, error) {
	log.Println("Listing devices for:", b.Name, b.RoomID)
	ch := make(chan []string)
	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain) + "+devices"
	eventType := event.MsgNotice
	eventSince := time.Now().UTC()
	eventSubscriber := EventSubscriber{
		Name:    eventSubName,
		MsgType: &eventType,
		Since:   &eventSince,
		RoomID:  b.RoomID,
		Callback: func(event *event.Event) {
			devicesRaw := strings.Split(event.Content.AsMessage().Body, "\n")
			devices := make([]string, 0)
			for _, device := range devicesRaw {
				deviceName, err := ExtractBracketContent(device)
				if err != nil {
					log.Println("Failed extracting device name", err, device)
					continue
				}
				devices = append(devices, deviceName)
			}
			ch <- devices

			defer func() {
				for index, subscriber := range EventSubscribers {
					if subscriber.Name == eventSubName {
						EventSubscribers = append(EventSubscribers[:index], EventSubscribers[index+1:]...)
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

func (b *Bridges) CreateContactRooms() error {
	log.Println("Joining member rooms for:", b.Name)

	clientDb := ClientDB{
		username: b.Client.UserID.Localpart(),
		filepath: "db/" + b.Client.UserID.Localpart() + ".db",
	}
	clientDb.Init()

	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain)
	eventSubName = eventSubName + "+join"

	processedRooms := make(map[id.RoomID]bool)

	eventSubscriber := EventSubscriber{
		Name:    eventSubName,
		MsgType: nil,
		ExcludeMsgTypes: []event.MessageType{
			event.MsgNotice, event.MsgVerificationRequest, event.MsgLocation,
		},
		Callback: func(evt *event.Event) {
			// log.Println("Received event:", event.RoomID, event.Content.AsMessage().Body)
			if evt.RoomID != "" {
				room := Rooms{
					Client: b.Client,
					ID:     evt.RoomID,
				}

				if _, ok := processedRooms[evt.RoomID]; ok {
					return
				}

				processedRooms[evt.RoomID] = true

				powerLevels, err := room.GetPowerLevelsUser()
				if err != nil {
					log.Println("Failed getting power levels", err)
					return
				}
				log.Println("Power levels:", powerLevels)
				powerLevelsEvents, err := room.GetPowerLevelsEvents()
				if err != nil {
					log.Println("Failed getting power levels events", err)
					return
				}
				log.Println("Power levels events:", powerLevelsEvents)

				isManagementRoom, err := room.IsManagementRoom(b.BotName)
				if err != nil {
					log.Println("Failed checking if room is management room", err)
					return
				}
				log.Println("Is management room:", evt.RoomID, isManagementRoom)
				processedRooms[evt.RoomID] = true

				if !isManagementRoom {
					members, err := room.GetRoomMembers(b.Client, evt.RoomID)
					if err != nil {
						log.Println("Failed getting room members", err)
						return
					}

					foundDevice := false
					foundMembers := make([]string, 0)
					foundDeviceUserName := ""

					for _, member := range members {
						log.Println("Checking member:", member.String())
						matched, err := cfg.CheckUsernameTemplate(b.Name, member.String())
						if err != nil {
							log.Println("Failed checking username template", err)
							return
						}
						if !matched {
							continue
						}

						devices := ClientDevices[b.Client.UserID.Localpart()][b.Name]
						log.Println("Devices:", devices)

						for _, device := range devices {
							formattedUsername, err := cfg.FormatUsername(b.Name, device)
							if err != nil {
								log.Println("Failed formatting username", err, device)
								continue
							}
							if member.String() == formattedUsername {
								foundDevice = true
								foundDeviceUserName = formattedUsername
								log.Println("Found device:", foundDeviceUserName)
								break
							} else {
								foundMembers = append(foundMembers, member.String())
							}
						}
					}

					if foundDevice && len(foundMembers) == 0 {
						log.Println("Found device but no members, adding device to members", foundDeviceUserName)
						foundMembers = append(foundMembers, foundDeviceUserName)
					}

					if foundDevice && len(foundMembers) > 0 {
						for _, fMember := range foundMembers {
							clientDb.StoreRooms(evt.RoomID.String(), b.Name, foundDeviceUserName, fMember, false)
							// log.Println("Stored room:", event.RoomID.String(), b.Name, fMember, false, foundDeviceUserName)
						}
					}
				}
			}
		},
	}

	EventSubscribers = append(EventSubscribers, eventSubscriber)

	return nil
}

func (b *Bridges) GetRoomInvitesDaemon() error {
	log.Println("Getting room invites for:", b.Name, b.RoomID)

	resp, err := b.Client.SyncRequest(context.Background(), 30000, "", "", true, event.PresenceOnline)
	if err != nil {
		log.Fatal(err)
	}

	for roomID := range resp.Rooms.Invite {
		log.Printf("You have been invited to room: %s\n", roomID)
		_, err := b.Client.JoinRoomByID(context.Background(), roomID)
		if err != nil {
			log.Println("Failed joining room", err)
		}
	}

	eventSubName := ReverseAliasForEventSubscriber(b.Client.UserID.Localpart(), b.Name, cfg.HomeServerDomain) + "+invites"
	eventSubscriber := EventSubscriber{
		Name:    eventSubName,
		MsgType: nil,
		Callback: func(evt *event.Event) {
			// log.Println("Received event:", evt.RoomID, evt.Content.AsMember())
			room := Rooms{
				Client: b.Client,
				ID:     evt.RoomID,
			}
			room.GetInvites(evt)
		},
	}

	EventSubscribers = append(EventSubscribers, eventSubscriber)

	return nil
}
