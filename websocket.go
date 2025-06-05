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

func GetWebsocketIndex(username string, platformName string) int {
	for index, _wd := range GlobalWebsocketConnection.Registry {
		if _wd.Username == username &&
			_wd.PlatformName == platformName {
			return index
		}
	}
	return -1
}

func (wd *WebsocketData) Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket handler called", wd.Bridge.Client.UserID)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(c *websocket.Conn) {
		defer wg.Done()
		for {
			data := <-wd.Bridge.ChImage
			fmt.Println(data)
			if data == nil {
				return
			}

			fmt.Println("Websocket sending message for:", wd.Bridge.Client.UserID)
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Websocket image"))

			if c == nil {
				log.Println("Error connecting socket, client is nil")
				return
			}

			err := c.WriteMessage(websocket.BinaryMessage, data)
			// err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error sending message to client socket", err)
				return
			}
		}
	}(conn)

	err = wd.Bridge.AddDevice(wd.Bridge.Client)
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		return
	}

	defer func() {
		// if index := GetWebsocketIndex(wd.Bridge.Client.UserID.Localpart(), wd.Bridge.Name); index > -1 {
		// 	GlobalWebsocketConnection.Registry =
		// 		slices.Delete(GlobalWebsocketConnection.Registry, index, index+1)
		// 	log.Println("Deleting websocket map at", index)
		// }
		log.Println("Falsifying syncing clients for user:", wd.Bridge.Client.UserID.Localpart())
		syncingClients.Registry[wd.Bridge.Client.UserID.Localpart()] = false
	}()

	wg.Wait()
}

func (wd *WebsocketData) RegisterWebsocket(platformName string, username string) {
	websocketUrl := fmt.Sprintf("/ws/%s/%s", platformName, username)
	if index := GetWebsocketIndex(username, platformName); index > -1 {
		log.Println("[+] Incoming socket connection but one already exist", wd.Bridge.Client.UserID)

		if !syncingClients.Registry[wd.Bridge.Client.UserID.Localpart()] {
			delete(syncingClients.Bridge, wd.Bridge.Client.UserID.Localpart())
			log.Println("Deleting syncing clients for user:", wd.Bridge.Client.UserID.Localpart())
			return
		}

		GlobalWebsocketConnection.Registry =
			slices.Delete(GlobalWebsocketConnection.Registry, index, index+1)
		wd.Bridge.Name = platformName
		log.Println("[+] Deleted socket at index", index)
	} else {
		http.HandleFunc(websocketUrl, wd.Handler)
		log.Println("[+] Registered websocket", websocketUrl)
	}
	GlobalWebsocketConnection.Registry = append(GlobalWebsocketConnection.Registry, &WebsocketMap{
		Url:          websocketUrl,
		PlatformName: platformName,
		Username:     username,
		Websocket:    wd,
	})
	wd.ch <- []byte(websocketUrl)
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
