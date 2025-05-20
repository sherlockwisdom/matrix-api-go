package main

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type BotsInterface interface {
	Invite()
	AddDevice(
		client *mautrix.Client,
		roomId string,
	)
	HandleMessages(*event.Event) (bool, error)
	ParseImage(*mautrix.Client, string) ([]byte, error)
}

type Bots struct{}
