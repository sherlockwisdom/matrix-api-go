package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "sherlock/matrix/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	// "maunium.net/go/mautrix/id"
)

// Users represents a user entity
// @Description Represents a user structure with a name
// @name Users
// @type object
type Users struct {
	Username    string `json:"username"`
	ID          int    `json:"id"`
	AccessToken string `json:"access_token"`
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

// LoginResponse represents the response for successful login
// @Description Response payload for successful login
type LoginResponse struct {
	Username    string `json:"username" example:""`
	AccessToken string `json:"access_token" example:""`
	Status      string `json:"status" example:""`
}

// ErrorResponse represents an error response
// @Description Response payload for error cases
type ErrorResponse struct {
	Error   string `json:"error" example:""`
	Details string `json:"details,omitempty" example:""`
}

// MessageResponse represents the response for successful message sending
// @Description Response payload for successful message sending
type MessageResponse struct {
	Contact string `json:"contact" example:""`
	EventID string `json:"event_id" example:""`
	Message string `json:"message" example:""`
	Status  string `json:"status" example:""`
}

// DeviceResponse represents the response for successful device addition
// @Description Response payload for successful device addition
type DeviceResponse struct {
	WebsocketURL string `json:"websocket_url" example:""`
}

// Input validation functions
func sanitizeUsername(username string) (string, error) {
	// Remove any whitespace
	username = strings.TrimSpace(username)

	// Username should be 3-32 characters and contain only letters, numbers, and underscores
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	if !validUsername.MatchString(username) {
		return "", fmt.Errorf("username must be 3-32 characters and contain only letters, numbers, and underscores")
	}

	return username, nil
}

func sanitizePassword(password string) (string, error) {
	// Remove any whitespace
	password = strings.TrimSpace(password)

	// Password should be at least 7 characters
	if len(password) < 7 {
		return "", fmt.Errorf("password must be at least 7 characters long")
	}

	return password, nil
}

func sanitizeMessage(message string) (string, error) {
	// Remove any whitespace
	message = strings.TrimSpace(message)

	// Message should not be empty and have a reasonable length
	if len(message) == 0 {
		return "", fmt.Errorf("message cannot be empty")
	}
	if len(message) > 4096 {
		return "", fmt.Errorf("message is too long (max 4096 characters)")
	}

	return message, nil
}

func sanitizePlatform(platform string) (string, error) {
	// Remove any whitespace and convert to lowercase
	platform = strings.ToLower(strings.TrimSpace(platform))

	// Platform should be 2-20 characters and contain only letters and numbers
	validPlatform := regexp.MustCompile(`^[a-z0-9]{2,20}$`)
	if !validPlatform.MatchString(platform) {
		return "", fmt.Errorf("platform name must be 2-20 characters and contain only letters and numbers")
	}

	return platform, nil
}

func sanitizeContact(contact string) (string, error) {
	// Remove any whitespace
	contact = strings.TrimSpace(contact)

	// E.164 format validation: +[country code][number], total length 8-15 digits
	validContact := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	if !validContact.MatchString(contact) {
		return "", fmt.Errorf("contact must be a valid E.164 phone number (e.g., +1234567890)")
	}

	return contact, nil
}

// ApiLogin godoc
// @Summary Logs a user into the Matrix server
// @Accept  json
// @Produce  json
// @Param   payload body ClientJsonRequest true "Login Credentials"
// @Success 200 {object} LoginResponse "Successfully logged in"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Login failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /login [post]
func ApiLogin(c *gin.Context) {
	var clientJsonRequest ClientJsonRequest

	if err := c.BindJSON(&clientJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Sanitize inputs
	username, err := sanitizeUsername(clientJsonRequest.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	password, err := sanitizePassword(clientJsonRequest.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
			User: Users{Username: username},
		},
	}

	if err := LoginProcess(client, &bridge, username, password); err != nil {
		log.Printf("Login failed for %s: %v", username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":     username,
		"access_token": client.AccessToken,
		"status":       "logged in",
	})
}

// ApiCreate godoc
// @Summary Creates a new user on the Matrix server
// @Accept  json
// @Produce  json
// @Param   payload body ClientJsonRequest true "User Registration"
// @Success 201 {object} LoginResponse "Successfully created user"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "User creation failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router / [post]
func ApiCreate(c *gin.Context) {
	var clientJsonRequest ClientJsonRequest

	if err := c.BindJSON(&clientJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Sanitize inputs
	username, err := sanitizeUsername(clientJsonRequest.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	password, err := sanitizePassword(clientJsonRequest.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
			User: Users{Username: username},
		},
	}

	if err := CreateProcess(client, &bridge, username, password); err != nil {
		log.Printf("User creation failed for %s: %v\n", username, err)
		c.JSON(http.StatusConflict, gin.H{"error": "User creation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"username":     username,
		"access_token": client.AccessToken,
		"status":       "created",
	})
}

// ApiSendMessage godoc
// @Summary Sends a message to a specified room
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name"
// @Param   contact path string true "Contact ID (E.164 phone number)"
// @Param   payload body ClientMessageJsonRequeset true "Message Payload"
// @Success 200 {object} MessageResponse "Message sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Failed to send message"
// @Router /{platform}/message/{contact} [post]
func ApiSendMessage(c *gin.Context) {
	var req ClientMessageJsonRequeset

	// Sanitize platform and contact parameters
	_, err := sanitizePlatform(c.Param("platform"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contactID, err := sanitizeContact(c.Param("contact"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.BindJSON(&req); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Sanitize message
	message, err := sanitizeMessage(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	room := Rooms{}

	resp, err := room.SendRoomMessages(client, message)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact":  contactID,
		"event_id": resp.EventID,
		"message":  message,
		"status":   "sent",
	})
}

// ApiAddDevice godoc
// @Summary Adds a device/bridge for a given platform
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name"
// @Param   payload body ClientBridgeJsonRequest true "Bridge Payload"
// @Success 200 {object} DeviceResponse "Successfully added device"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Bridge not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /{platform}/devices/ [post]
func ApiAddDevice(c *gin.Context) {
	var bridgeJsonRequest ClientBridgeJsonRequest

	// Sanitize platform parameter
	platformName, err := sanitizePlatform(c.Param("platform"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&bridgeJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Sanitize username
	username, err := sanitizeUsername(bridgeJsonRequest.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var bridge *Bridges
	for _, _bridge := range syncingClients.Bridge[username] {
		log.Println("[Add Device] Checking Bridge for user:", username, _bridge.Name)
		if _bridge.Name == platformName {
			bridge = _bridge
			break
		}
	}

	if bridge == nil {
		log.Println("Bridge not found for user:", username)
		c.JSON(http.StatusNotFound, gin.H{"error": "Bridge not found"})
		return
	}

	var websocket = WebsocketData{
		ch:     make(chan []byte, 1),
		Bridge: bridge,
	}

	client, err := mautrix.NewClient(
		cfg.HomeServer,
		id.NewUserID(username, cfg.HomeServerDomain),
		bridgeJsonRequest.AccessToken,
	)

	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	ProcessActiveSessions(client, username, "", "", bridge, true)

	websocket.RegisterWebsocket(platformName, username)

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

func CliFlow() {
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

// @title           ShortMesh API
// @version         1.0
// @description     ShortMesh is a Matrix-based messaging bridge API that enables seamless communication across different messaging platforms. It provides endpoints for user management, message sending, and platform bridging capabilities. The API supports E.164 phone number format for contacts and implements secure authentication mechanisms.
// @host      localhost:8080
func main() {
	if cfgError != nil {
		panic(cfgError)
	}

	if len(os.Args) > 1 {
		CliFlow()
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.POST("/", ApiCreate)
	router.POST("/login", ApiLogin)
	router.POST("/:platform/message/:contact", ApiSendMessage)
	router.POST("/:platform/devices/", ApiAddDevice)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	ks.Init()

	host := cfg.Server.Host
	port := cfg.Server.Port
	tlsCert := cfg.Server.Tls.Crt
	tlsKey := cfg.Server.Tls.Key

	if tlsCert != "" && tlsKey != "" {
		go func() {
			err := GlobalWebsocketConnection.MainWebsocket(true)
			if err != nil {
				panic(err)
			}
		}()
		router.RunTLS(fmt.Sprintf(":%s", port), tlsCert, tlsKey)
		return
	}

	go func() {
		err := GlobalWebsocketConnection.MainWebsocket(false)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := SyncAllClients()
		if err != nil {
			panic(err)
		}
	}()

	router.Run(fmt.Sprintf("%s:%s", host, port))
}
