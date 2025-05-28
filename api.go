package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Users struct {
	name string
}

type ClientJsonRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientMessageJsonRequeset struct {
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}

type ClientBridgeJsonRequest struct {
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
}

func ApiLogin(c *gin.Context) {
	var clientJsonRequest ClientJsonRequest

	if err := c.BindJSON(&clientJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if clientJsonRequest.Username == "" || clientJsonRequest.Password == "" {
		log.Println("Missing username or password in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var bridge = Bridges{
		ch: make(chan *event.Event),
		room: Rooms{
			User: Users{name: clientJsonRequest.Username},
		},
	}

	if err := LoginProcess(client, &bridge, clientJsonRequest.Username, clientJsonRequest.Password); err != nil {
		log.Printf("Login failed for %s: %v", clientJsonRequest.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":     clientJsonRequest.Username,
		"access_token": client.AccessToken,
		"status":       "logged in",
	})
}

func ApiCreate(c *gin.Context) {
	var clientJsonRequest ClientJsonRequest

	if err := c.BindJSON(&clientJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Check that required fields are not empty
	if clientJsonRequest.Username == "" || clientJsonRequest.Password == "" {
		log.Println("Missing username or password in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var bridge = Bridges{
		ch: make(chan *event.Event),
		room: Rooms{
			User: Users{name: clientJsonRequest.Username},
		},
	}

	if err := CreateProcess(client, &bridge, clientJsonRequest.Username, clientJsonRequest.Password); err != nil {
		log.Printf("User creation failed for %s: %v", clientJsonRequest.Username, err)
		c.JSON(http.StatusConflict, gin.H{"error": "User creation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"username": clientJsonRequest.Username,
		"status":   "created",
	})
}

func ApiSendMessage(c *gin.Context) {
	var req ClientMessageJsonRequeset
	roomID := c.Param("roomid")

	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room ID"})
		return
	}

	if err := c.BindJSON(&req); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message body cannot be empty"})
		return
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer

	client, err := mautrix.NewClient(homeServer, "", req.AccessToken)
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	room := Rooms{
		ID: id.RoomID(roomID),
	}

	resp, err := room.SendRoomMessages(client, req.Message)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id":  roomID,
		"event_id": resp.EventID,
		"message":  req.Message,
		"status":   "sent",
	})
}

func ApiAddDevice(c *gin.Context) {
	var bridgeJsonRequest ClientBridgeJsonRequest
	platformName := c.Param("platform")
	log.Println("API request platform name:", platformName)

	if err := c.ShouldBindJSON(&bridgeJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	bridge := Bridges{
		ch:   make(chan *event.Event, 1),
		name: platformName,
		room: Rooms{
			User: Users{bridgeJsonRequest.Username},
		},
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer
	client, err := mautrix.NewClient(
		homeServer,
		id.UserID(fmt.Sprintf("@%s:%s", bridgeJsonRequest.Username, cfg.HomeServerDomain)),
		bridgeJsonRequest.AccessToken,
	)
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Matrix client"})
		return
	}

	image, err := bridge.AddDevice(client)
	if err != nil {
		log.Printf("Failed to add device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add device"})
		return
	}

	var websocket = WebsocketData{
		ch:    make(chan []byte, 1),
		image: image,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		websocketUrl := <-websocket.ch

		c.JSON(http.StatusOK, gin.H{
			"websocket_url": string(websocketUrl),
		})
		defer wg.Done()
	}()

	go func() {
		err = websocket.MainWebsocket(platformName, bridgeJsonRequest.Username)
		if err != nil {
			log.Println("Failed to start websocket:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start websocket"})
			return
		}
	}()

	wg.Wait()
}

func main() {
	if len(os.Args) > 1 {
		cfg, _ := (&Conf{}).getConf()
		homeServer := cfg.HomeServer
		password := "M4yHFt$5hW0UuyTv2hdRwtGryHa9$R7z"

		client, err := mautrix.NewClient(homeServer, "", "")
		if err != nil {
			panic(err)
		}

		var bridge = Bridges{
			ch: make(chan *event.Event, 500),
		}
		switch os.Args[1] {
		case "--create":
			username := "sherlock_" + strconv.FormatInt(time.Now().UnixMilli(), 10)
			err := CreateProcess(
				client,
				&bridge,
				username,
				password,
			)

			if err != nil {
				panic(err)
			}
		case "--login":
			username := os.Args[2]
			LoginProcess(client, &bridge, username, password)
		case "--websocket":
			var wd = WebsocketData{ch: make(chan []byte, 1)}
			wd.ch <- []byte("may the force!")
			err := wd.MainWebsocket("testingPlatform", "testingUser")
			if err != nil {
				panic(err)
			}
			os.Exit(0)
		default:
		}
		CompleteRun(client, &bridge)
	}

	router := gin.Default()
	router.POST("/", ApiCreate)
	router.POST("/login", ApiLogin)
	router.POST("/:platform/message/:roomid", ApiSendMessage)
	router.POST("/:platform/devices/", ApiAddDevice)

	router.Run("localhost:8080")
}
