package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

func (ws *Websockets) listenForDisconnection(c *websocket.Conn, ch chan []byte) {
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
				ch <- nil
			}
		}
	}
}

func (ws *Websockets) Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket handler called", ws.Bridge.Client.UserID)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	ch := make(chan []byte)
	go ws.listenForDisconnection(conn, ch)

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

	err = ws.Bridge.AddDevice(ch)
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

	for {
		log.Println("Waiting for data from channel")
		data := <-ch
		if data == nil {
			err := conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Printf("Error sending message to client socket for user %s: %v", ws.Bridge.Client.UserID, err)
			}
			conn.Close()
			break
		}

		fmt.Println("Websocket sending message for:", ws.Bridge.Client.UserID)

		if conn == nil {
			log.Println("Error connecting socket, client is nil")
			break
		}

		err := conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Printf("Error sending message to client socket for user %s: %v", ws.Bridge.Client.UserID, err)
			break
		}
	}

	defer func() {
		eventSubName := ReverseAliasForEventSubscriber(ws.Bridge.Client.UserID.Localpart(), ws.Bridge.Name, cfg.HomeServerDomain)
		for index, subscriber := range EventSubscribers {
			if subscriber.Name == eventSubName {
				EventSubscribers = append(EventSubscribers[:index], EventSubscribers[index+1:]...)
				log.Println("Removed event subscriber:", eventSubName)
				break
			}
		}
	}()
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
