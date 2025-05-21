package main

import (
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type BridgesInterface interface {
	Invite()
	AddDevice(
		client *mautrix.Client,
		roomId string,
	)
	HandleMessages(*event.Event) (bool, error)
	ParseImage(*mautrix.Client, string) ([]byte, error)
}

type Bridges struct {
	username string
}

func (bridge *Bridges) HandleMessage(evt *event.Event) (bool, error) {
	// check room
	// check template

	if evt.Type == event.EventMessage {
		var clientDB ClientDB = ClientDB{
			username: bridge.username,
			filepath: "db/" + bridge.username + ".db",
		}

		clientDB.Init()
		defer clientDB.Close()

		room, err := clientDB.FetchRooms(evt.RoomID.String())

		if err != nil {
			return false, err
		}

		if !room.isBridge {
			return false, nil
		}

		log.Println("[+] BRIDGE| New message:", evt.Content.AsMessage().Body)
		return true, nil
	}
	return false, nil
}
