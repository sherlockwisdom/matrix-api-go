package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type ClientJsonRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientMessageJsonRequeset struct {
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
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

	homeServer := "https://relaysms.me"

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	bridge := Bridges{username: clientJsonRequest.Username}
	room := Rooms{
		Channel: make(chan *event.Event),
		Bridge:  bridge,
	}

	if err := LoginProcess(client, &room, clientJsonRequest.Username, clientJsonRequest.Password); err != nil {
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

	homeServer := "https://relaysms.me"

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var bridge = Bridges{username: clientJsonRequest.Username}
	room := Rooms{
		Channel: make(chan *event.Event),
		Bridge:  bridge,
	}

	if err := CreateProcess(client, &room, clientJsonRequest.Username, clientJsonRequest.Password); err != nil {
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

	homeServer := "https://relaysms.me"
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

func main() {
	if len(os.Args) > 1 {
		password := "M4yHFt$5hW0UuyTv2hdRwtGryHa9$R7z"
		homeServer := "https://relaysms.me"

		client, err := mautrix.NewClient(homeServer, "", "")
		if err != nil {
			panic(err)
		}

		var bridge Bridges
		var room = Rooms{
			Channel: make(chan *event.Event, 500),
			Bridge:  bridge,
		}
		switch os.Args[1] {
		case "--create":
			username := "sherlock_" + strconv.FormatInt(time.Now().UnixMilli(), 10)
			CreateProcess(
				client,
				&room,
				username,
				password,
			)
		case "--login":
			username := os.Args[2]
			LoginProcess(client, &room, username, password)
		default:
		}
		CompleteRun(client, &room)
	}

	router := gin.Default()
	router.POST("/create", ApiCreate)
	router.POST("/login", ApiLogin)
	router.POST("/:platform/message/:roomid", ApiSendMessage)

	router.Run("localhost:8080")
}
