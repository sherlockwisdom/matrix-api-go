package main

import (
	"context"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func JoinedRooms(client *mautrix.Client) ([]id.RoomID, error) {
	resp, err := client.JoinedRooms(context.Background())

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
	}

	return resp.JoinedRooms, err
}
