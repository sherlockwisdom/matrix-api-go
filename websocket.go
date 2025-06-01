package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type WebsocketDataInterface interface {
	Handler(http.ResponseWriter, *http.Request)
}

type WebsocketData struct {
	ch     chan []byte
	Bridge *Bridges
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	err := conn.WriteMessage(websocket.TextMessage, []byte("Welcome!!"))

	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
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
			err := conn.WriteMessage(websocket.TextMessage, []byte("Websocket image"))
			// err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error sending message to client socket", err)
			}
		}
	}()

	err = wd.Bridge.AddDevice(wd.Bridge.Client)
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

	wg.Wait()
}

func (wd *WebsocketData) MainWebsocket(platformName string, username string) error {
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)
	http.HandleFunc(websocketUrl, wd.Handler)
	wd.ch <- []byte(websocketUrl)

	return http.ListenAndServe(":8090", nil)
}
