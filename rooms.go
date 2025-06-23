package main

import (
	"context"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Rooms struct {
	Client     *mautrix.Client
	ID         id.RoomID
	isBridge   bool
	DeviceName string
	Members    map[string]string
}

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

	isSpace, err := r.IsSpaceRoom()
	if err != nil {
		return false, err
	}

	if !isSpace {
		if len(members.Joined) == 2 {
			botID := id.UserID(botName)
			if _, ok := members.Joined[botID]; ok {
				if _, ok := members.Joined[r.Client.UserID]; ok {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (r *Rooms) GetRoomInfo() (string, error) {
	// Get room name
	var nameContent event.RoomNameEventContent
	err := r.Client.StateEvent(context.Background(), r.ID, event.StateRoomName, "", &nameContent)
	if err != nil {
		return "", err
	}

	return nameContent.Name, nil
}

// IsSpaceRoom checks if the given room is a space
func (r *Rooms) IsSpaceRoom() (bool, error) {
	var createContent event.CreateEventContent

	err := r.Client.StateEvent(context.Background(), r.ID, event.StateCreate, "", &createContent)
	if err != nil {
		return false, err
	}

	// Check if "type" field is set to "m.space"
	if createContent.Type == "m.space" {
		return true, nil
	}
	return false, nil
}
