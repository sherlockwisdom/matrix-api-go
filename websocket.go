package main

import (
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

type WebsocketDataInterface interface {
	Handler(http.ResponseWriter, *http.Request)
}

type WebsocketData struct {
	ch     chan []byte
	Bridge *Bridges
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(c *websocket.Conn) {
		for {
			data := <-wd.Bridge.ChImage
			// msgType, msg, err := conn.ReadMessage()
			// if err != nil {
			// 	break
			// }
			// // Print received message
			// println("Received:", string(msg))

			// // Respond with Hello, World!
			fmt.Println(data)
			fmt.Println("Websocket starting for:", wd.Bridge.Client.UserID)
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Websocket image"))

			if c == nil {
				log.Println("Error connecting socket, client is nil")
				return
			}

			err := c.WriteMessage(websocket.BinaryMessage, data)
			// err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error sending message to client socket", err)
			}
		}
	}(conn)

	err = wd.Bridge.AddDevice(wd.Bridge.Client)
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

	wg.Wait()
}

func (wd *WebsocketData) MainWebsocket(platformName string, username string) error {
	wd.Bridge.Room.User.name = username
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)
	http.HandleFunc(websocketUrl, wd.Handler)
	wd.ch <- []byte(websocketUrl)

	cfg, err := (&Conf{}).getConf()
	if err != nil {
		panic(err)
	}

	if cfg.Server.Tls.Crt != "" && cfg.Server.Tls.Key != "" {
		log.Println("Starting websocket with Tls")
		return http.ListenAndServeTLS(":8090", cfg.Server.Tls.Crt, cfg.Server.Tls.Key, nil)
	}

	log.Println("Starting websocket without Tls")
	return http.ListenAndServe(":8090", nil)
}
