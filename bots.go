package main

type BotsInterface interface {
	Invite()
	AddDevice()
}

type Bots struct{}
