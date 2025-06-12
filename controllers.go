package main

import (
	"log"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type Controller struct {
	Client   *mautrix.Client
	Username string
	UserID   id.UserID
}

type UserSync struct {
	Name         string
	MsgBridges   []*Bridges
	LoginBridges []*Bridges
	Syncing      bool
	SyncMutex    sync.Mutex
}

var cfg, cfgError = (&Conf{}).getConf()

var GlobalWebsocketConnection = WebsocketController{
	Registry: make([]*WebsocketUnit, 0),
}

var GlobalController = Controller{
	Client: &mautrix.Client{
		UserID:      id.NewUserID(cfg.User.Username, cfg.HomeServerDomain),
		AccessToken: cfg.User.AccessToken,
	},
	Username: cfg.User.Username,
}

var ks = Keystore{
	filepath: cfg.KeystoreFilepath,
}

var syncingClients = SyncingClients{
	Users: make(map[string]*UserSync),
}

func (c *Controller) CreateProcess(password string) error {
	m := MatrixClient{
		Client: c.Client,
	}
	accessToken, err := m.Create(c.Username, password)

	if err != nil {
		return err
	}

	c.Client.UserID = id.NewUserID(c.Username, cfg.HomeServerDomain)
	c.Client.AccessToken = accessToken
	log.Println("[+] Created user: ", c.Username)

	err = m.ProcessActiveSessions(password)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) LoginProcess(password string) error {
	m := MatrixClient{
		Client: c.Client,
	}
	accessToken, err := m.LoadActiveSessions(password)
	if err != nil {
		return err
	}

	if accessToken == "" {
		if accessToken, err = m.Login(password); err != nil {
			return err
		}
	}

	c.Client.AccessToken = accessToken
	err = m.ProcessActiveSessions(password)
	if err != nil {
		return err
	}

	return nil
}
