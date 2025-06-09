package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// var upgrader = websocket.Upgrader{}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development/testing.
		// In production, you should restrict this to known origins.
		return true
	},
}

type Websockets struct {
	Bridge *Bridges
}

type WebsocketController struct {
	Registry []*WebsocketUnit
}

type WebsocketUnit struct {
	Url          string
	PlatformName string
	Username     string
	Websocket    *Websockets
}

func GetWebsocketUsernameIndex(username string) int {
	for index, _wd := range GlobalWebsocketConnection.Registry {
		if _wd.Username == username {
			return index
		}
	}
	return -1
}

func GetWebsocketIndex(username string, platformName string) int {
	for index, _wd := range GlobalWebsocketConnection.Registry {
		if _wd.Username == username &&
			_wd.PlatformName == platformName {
			return index
		}
	}
	return -1
}

func (ws *Websockets) Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket handler called", ws.Bridge.Client.UserID)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	syncingClients.Users[ws.Bridge.Client.UserID.Localpart()].Bridges = append(
		syncingClients.Users[ws.Bridge.Client.UserID.Localpart()].Bridges,
		ws.Bridge,
	)

	var wg sync.WaitGroup
	wg.Add(2) // Increased to 2 for the new goroutine

	// Add connection monitoring goroutine
	go func(c *websocket.Conn) {
		defer wg.Done()
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Printf("Client connection lost for user %s: %v", ws.Bridge.Client.UserID, err)
				if bridgeCfg, ok := cfg.GetBridgeConfig(ws.Bridge.Name); ok {
					log.Println("Sending cancel command to:", ws.Bridge.RoomID)
					_, err = ws.Bridge.Client.SendText(
						context.Background(),
						ws.Bridge.RoomID,
						bridgeCfg.Cmd["cancel"],
					)
				}
				break
			}
		}
	}(conn)

	go func(c *websocket.Conn) {
		defer wg.Done()
		for {
			data := <-ws.Bridge.ChImageSyncEvt
			if data == nil {
				c.Close()
				return
			}

			fmt.Println("Websocket sending message for:", ws.Bridge.Client.UserID)

			if c == nil {
				log.Println("Error connecting socket, client is nil")
				return
			}

			err := c.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Printf("Error sending message to client socket for user %s: %v", ws.Bridge.Client.UserID, err)
				return
			}
		}
	}(conn)

	err = ws.Bridge.AddDevice()
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

	defer func() {
		log.Println("Falsifying syncing clients for user:", ws.Bridge.Client.UserID.Localpart())
		userSync := syncingClients.Users[ws.Bridge.Client.UserID.Localpart()]
		if userSync != nil {
			// remove bridge from syncing clients
			for index, bridge := range userSync.Bridges {
				if bridge.Name == ws.Bridge.Name {
					userSync.Bridges = append(userSync.Bridges[:index], userSync.Bridges[index+1:]...)
					break
				}
			}
		}
	}()

	wg.Wait()
}

func (w *Websockets) RegisterWebsocket(platformName string, username string) string {
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)

	http.HandleFunc(websocketUrl, w.Handler)
	log.Println("[+] Registered websocket", websocketUrl)
	GlobalWebsocketConnection.Registry = append(GlobalWebsocketConnection.Registry, &WebsocketUnit{
		Url:          websocketUrl,
		PlatformName: platformName,
		Username:     username,
		Websocket:    w,
	})
	return websocketUrl
}

func MainWebsocket(tls bool) error {
	cfg, err := (&Conf{}).getConf()
	if err != nil {
		panic(err)
	}

	if tls {
		log.Println("Starting websocket with Tls")
		return http.ListenAndServeTLS(":8090", cfg.Server.Tls.Crt, cfg.Server.Tls.Key, nil)
	}

	log.Println("Starting websocket without Tls")
	return http.ListenAndServe(":8090", nil)
}
