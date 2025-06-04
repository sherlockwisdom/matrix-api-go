package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "sherlock/matrix/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var GlobalWebsocketConnection = WebsocketData{
	ch: make(chan []byte, 500),
}

// Users represents a user entity
// @Description Represents a user structure with a name
// @name Users
// @type object
type Users struct {
	name string
}

// ClientJsonRequest represents login or registration data
// @Description Request payload for user login or registration
// @name ClientJsonRequest
// @type object
type ClientJsonRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ClientMessageJsonRequeset represents a message sending request
// @Description Request payload to send a message to a room
// @name ClientMessageJsonRequeset
// @type object
type ClientMessageJsonRequeset struct {
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}

// ClientBridgeJsonRequest represents bridge connection details
// @Description Request payload to bind a platform bridge to a user
// @name ClientBridgeJsonRequest
// @type object
type ClientBridgeJsonRequest struct {
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
}

// ApiLogin godoc
// @Summary Logs a user into the Matrix server
// @Accept  json
// @Produce  json
// @Param   payload body ClientJsonRequest true "Login Credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
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
		ChEvt: make(chan *event.Event),
		Room: Rooms{
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

// ApiCreate godoc
// @Summary Creates a new user on the Matrix server
// @Accept  json
// @Produce  json
// @Param   payload body ClientJsonRequest true "User Registration"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router / [post]
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
		ChEvt: make(chan *event.Event),
		Room: Rooms{
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

// ApiSendMessage godoc
// @Summary Sends a message to a specified room
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name"
// @Param   roomid path string true "Room ID"
// @Param   payload body ClientMessageJsonRequeset true "Message Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /{platform}/message/{roomid} [post]
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

// ApiAddDevice godoc
// @Summary Adds a device/bridge for a given platform
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name"
// @Param   payload body ClientBridgeJsonRequest true "Bridge Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /{platform}/devices/ [post]
func ApiAddDevice(c *gin.Context) {
	var bridgeJsonRequest ClientBridgeJsonRequest
	platformName := c.Param("platform")
	log.Println("API request platform name:", platformName)

	if err := c.ShouldBindJSON(&bridgeJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer

	client, err := mautrix.NewClient(
		homeServer,
		// id.UserID(fmt.Sprintf("@%s:%s", bridgeJsonRequest.Username, cfg.HomeServerDomain)),
		id.NewUserID(bridgeJsonRequest.Username, cfg.HomeServerDomain),
		bridgeJsonRequest.AccessToken,
	)

	bridge := Bridges{
		ChEvt:   make(chan *event.Event, 1),
		ChImage: make(chan []byte, 1),
		Name:    platformName,
		Room: Rooms{
			User: Users{bridgeJsonRequest.Username},
		},
		Client: client,
	}

	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Matrix client"})
		return
	}

	var websocket = WebsocketData{
		ch:     make(chan []byte, 1),
		Bridge: &bridge,
	}

	websocket.RegisterWebsocket(platformName, bridgeJsonRequest.Username)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		websocketUrl := <-websocket.ch

		c.JSON(http.StatusOK, gin.H{
			"websocket_url": string(websocketUrl),
		})
		defer wg.Done()
	}()

	wg.Wait()
}

// @title           Swagger Example API
// @version         2.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
			ChEvt: make(chan *event.Event, 500),
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
			err := wd.MainWebsocket(false)
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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	cfg, err := (&Conf{}).getConf()
	if err != nil {
		panic(err)
	}

	host := cfg.Server.Host
	port := cfg.Server.Port
	if cfg.Server.Tls.Crt != "" && cfg.Server.Tls.Key != "" {
		go func() {
			err := GlobalWebsocketConnection.MainWebsocket(true)
			if err != nil {
				panic(err)
			}
		}()
		router.RunTLS(fmt.Sprintf(":%s", port), cfg.Server.Tls.Crt, cfg.Server.Tls.Key)
		return
	}

	go func() {
		err := GlobalWebsocketConnection.MainWebsocket(false)
		if err != nil {
			panic(err)
		}
	}()
	router.Run(fmt.Sprintf("%s:%s", host, port))
}
