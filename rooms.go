package main

import (
	"context"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Rooms struct {
	Client   *mautrix.Client
	ID       id.RoomID
	isBridge bool
	Members  map[string]string
}

// func (r *Rooms) Join(
// 	client *mautrix.Client,
// 	roomId id.RoomID,
// ) error {
// 	log.Println("[*] Joining room:", roomId)
// 	_, err := client.JoinRoomByID(context.Background(), roomId)
// 	return err
// }

func (r *Rooms) IsBridgeInviteForContact(evt *event.Event) (bool, error) {
	// TODO: check if the invite is from a bridge bot but not a bridge room
	for _, bridge := range cfg.Bridges {
		for _, bridgeCfg := range bridge {
			if bridgeCfg.BotName == evt.Sender.String() {
				isBridge, err := r.IsBridgeMessage(evt)
				if err != nil {
					return false, err
				}
				return !isBridge, nil
			}
		}
	}

	return false, nil
}

func (r *Rooms) IsBridgeMessage(evt *event.Event) (bool, error) {
	if evt.Type == event.EventMessage {
		var clientDB ClientDB = ClientDB{
			username: r.Client.UserID.Localpart(),
			filepath: "db/" + r.Client.UserID.Localpart() + ".db",
		}

		clientDB.Init()
		defer clientDB.Close()

		room, err := clientDB.FetchRooms(evt.RoomID.String())

		if err != nil {
			return false, err
		}

		return room.isBridge, nil
	}
	return false, nil
}

func (r *Rooms) GetRoomMembers(client *mautrix.Client, roomId id.RoomID) ([]id.UserID, error) {
	members, err := client.JoinedMembers(context.Background(), roomId)

	if err != nil {
		return nil, err
	}

	var membersList []id.UserID
	for userId, _ := range members.Joined {
		membersList = append(membersList, userId)
	}

	return membersList, nil
}

func (r *Rooms) IsManagementRoom(botName string) (bool, error) {
	members, err := r.Client.JoinedMembers(context.Background(), r.ID)
	if err != nil {
		return false, err
	}

	if len(members.Joined) == 2 {
		for userID, _ := range members.Joined {
			if userID.String() == botName {
				return true, nil
			}
		}
	}

	return false, nil
}
