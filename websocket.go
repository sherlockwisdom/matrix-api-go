package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
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
	ch       chan []byte
	Bridge   *Bridges
	Registry []*WebsocketMap
}

type WebsocketMap struct {
	Url          string
	PlatformName string
	Username     string
	Websocket    *WebsocketData
}

func GetWebsocketIndex(wd *WebsocketData) int {
	for index, _wd := range GlobalWebsocketConnection.Registry {
		if _wd.Websocket == wd {
			return index
		}
	}
	return -1
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	if index := GetWebsocketIndex(wd); index == -1 {
		log.Println("[+] Incoming socket connection but no mapped request", wd.Bridge.Client.UserID)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(c *websocket.Conn) {
		for {
			data := <-wd.Bridge.ChImage
			fmt.Println(data)
			if data == nil {
				if index := GetWebsocketIndex(wd); index > -1 {
					GlobalWebsocketConnection.Registry =
						slices.Delete(GlobalWebsocketConnection.Registry, index, index+1)
					log.Println("Deleting websocket map at", index)
					wg.Done()
				}
			}
			fmt.Println("Websocket starting for:", wd.Bridge.Client.UserID)
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Websocket image"))

			if c == nil {
				log.Println("Error connecting socket, client is nil")
				wg.Done()
			}

			err := c.WriteMessage(websocket.BinaryMessage, data)
			// err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error sending message to client socket", err)
				if index := GetWebsocketIndex(wd); index > -1 {
					GlobalWebsocketConnection.Registry =
						slices.Delete(GlobalWebsocketConnection.Registry, index, index+1)
					log.Println("Deleting websocket map at", index)
				}
				wg.Done()
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

func (wd *WebsocketData) RegisterWebsocket(platformName string, username string) {
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)
	http.HandleFunc(websocketUrl, wd.Handler)
	GlobalWebsocketConnection.Registry = append(GlobalWebsocketConnection.Registry, &WebsocketMap{
		Url:          websocketUrl,
		PlatformName: platformName,
		Username:     username,
		Websocket:    wd,
	})
	wd.ch <- []byte(websocketUrl)
	log.Println("[+] Registered websocket", websocketUrl)
}

func (wd *WebsocketData) MainWebsocket(tls bool) error {
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
