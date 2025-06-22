package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"maunium.net/go/mautrix/event"
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

	ws.Bridge.ChLoginSyncEvt = make(chan *event.Event)

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

					if err != nil {
						log.Printf("Error sending cancel command to %s: %v", ws.Bridge.RoomID, err)
					}
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
				err := c.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					log.Printf("Error sending message to client socket for user %s: %v", ws.Bridge.Client.UserID, err)
				}
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

	clientDb := ClientDB{
		username: ws.Bridge.Client.UserID.Localpart(),
		filepath: "db/" + ws.Bridge.Client.UserID.Localpart() + ".db",
	}
	if err := clientDb.Init(); err != nil {
		log.Println("Error initializing client db:", err)
	}

	if IsActiveSessionsExpired(&clientDb, ws.Bridge.Client.UserID.Localpart()) {
		log.Println("Active sessions expired, removing active sessions")
		clientDb.RemoveActiveSessions(ws.Bridge.Client.UserID.Localpart())
	} else {
		sessions, _, err := clientDb.FetchActiveSessions(ws.Bridge.Client.UserID.Localpart())
		if err != nil {
			log.Println("Error fetching active sessions:", err)
		}

		if len(sessions) > 0 {
			log.Println("Active sessions found, sending message to client socket")
		}
		conn.WriteMessage(websocket.BinaryMessage, sessions)
	}

	err = ws.Bridge.AddDevice()
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

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
	port := cfg.Websocket.Port
	host := cfg.Websocket.Host

	if tls {
		log.Println("Starting websocket with Tls")
		return http.ListenAndServeTLS(fmt.Sprintf("%s:%s", host, port), cfg.Websocket.Tls.Crt, cfg.Websocket.Tls.Key, nil)
	}

	log.Println("Starting websocket without Tls")
	return http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil)
}
