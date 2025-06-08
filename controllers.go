package main

import (
	"log"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type Controller struct {
	Client      *mautrix.Client
	Username    string
	Password    string
	AccessToken string
}

var cfg, cfgError = (&Conf{}).getConf()
var GlobalWebsocketConnection = WebsocketData{
	ch:       make(chan []byte, 500),
	Registry: make([]*WebsocketMap, 0),
}

var ks = Keystore{
	filepath: cfg.KeystoreFilepath,
}

var (
	syncingClients = SyncingClients{
		Bridge:   make(map[string][]*Bridges),
		Registry: make(map[string]bool),
	}
	mapMutex = sync.Mutex{}
)

func (c *Controller) CreateProcess() error {
	m := MatrixClient{
		Client: c.Client,
	}
	accessToken, err := m.Create(c.Username, c.Password)

	if err != nil {
		return err
	}

	c.Client.UserID = id.NewUserID(c.Username, cfg.HomeServerDomain)
	c.Client.AccessToken = accessToken
	log.Println("[+] Created user: ", c.Username)

	err = m.ProcessActiveSessions(c.Password)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) LoginProcess() error {
	m := MatrixClient{
		Client: c.Client,
	}
	accessToken, err := m.LoadActiveSessions(c.Password)
	if err != nil {
		return err
	}

	if accessToken == "" {
		if accessToken, err = m.Login(c.Username, c.Password); err != nil {
			return err
		}
	}

	c.Client.AccessToken = accessToken
	err = m.ProcessActiveSessions(c.Password)
	if err != nil {
		return err
	}

	return nil
}
