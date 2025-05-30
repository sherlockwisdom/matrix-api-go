package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type WebsocketDataInterface interface {
	Handler(http.ResponseWriter, *http.Request)
}

type WebsocketData struct {
	ch     chan []byte
	Bridge Bridges
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	err := conn.WriteMessage(websocket.TextMessage, []byte("Welcome!!"))

	if err != nil {
		log.Println(err)
	}

	for {
		// msgType, msg, err := conn.ReadMessage()
		// if err != nil {
		// 	break
		// }
		// // Print received message
		// println("Received:", string(msg))

		// // Respond with Hello, World!
		data := <-wd.Bridge.chImage
		fmt.Println(data)
		err := conn.WriteMessage(websocket.TextMessage, []byte("Websocket image"))
		if err != nil {
			log.Println(err)
		}
	}
}

func (wd *WebsocketData) MainWebsocket(platformName string, username string) error {
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)
	http.HandleFunc(websocketUrl, wd.Handler)
	wd.ch <- []byte(websocketUrl)
	return http.ListenAndServe(":8090", nil)
}
