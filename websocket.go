package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type WebsocketDataInterface interface {
	Handler(http.ResponseWriter, *http.Request)
}

type WebsocketData struct {
	ch    *chan []byte
	image []byte
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	for {
		// msgType, msg, err := conn.ReadMessage()
		// if err != nil {
		// 	break
		// }
		// // Print received message
		// println("Received:", string(msg))

		// // Respond with Hello, World!
		data := <-*wd.ch
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (wd *WebsocketData) MainWebsocket() error {
	http.HandleFunc("/ws", wd.Handler)
	return http.ListenAndServe(":8090", nil)
}
