package main

import "maunium.net/go/mautrix/event"

type BotsInterface interface {
	Invite()
	AddDevice()
}

type Bots struct {
	channel chan *event.Event
}
