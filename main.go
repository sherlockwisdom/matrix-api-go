package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "sherlock/matrix/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	// "maunium.net/go/mautrix/id"
)

// @title           ShortMesh API
// @version         1.0
// @description     ShortMesh is a Matrix-based messaging bridge API that enables seamless communication across different messaging platforms.
// @description     It provides endpoints for user management, message sending, and platform bridging capabilities.
// @description     The API supports E.164 phone number format for contacts and implements secure authentication mechanisms.
// @description     The API supports the following platforms:
// @description     - WhatsApp
// @description     - Signal (coming soon)
// @host      localhost:8080
// @BasePath  /
// @schemes   http https

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
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"securepassword123"`
}

// ClientMessageJsonRequeset represents a message sending request
// @Description Request payload to send a message to a contact through a platform bridge. All fields are validated according to specific rules.
// @name ClientMessageJsonRequeset
// @type object
type ClientMessageJsonRequeset struct {
	Username   string `json:"username" example:"john_doe" binding:"required"`       // Required: 3-32 characters, letters, numbers, underscores only
	Message    string `json:"message" example:"Hello, world!" binding:"required"`   // Required: 1-4096 characters, cannot be empty
	DeviceName string `json:"device_name" example:"wa123456789" binding:"required"` // Required: 2-20 characters, letters and numbers only
	FileData   []byte `json:"file_data,omitempty" example:"[file_data]"`            // Optional: Binary file data for attachments
}

// ClientBridgeJsonRequest represents bridge connection details
// @Description Request payload to bind a platform bridge to a user
// @name ClientBridgeJsonRequest
// @type object
type ClientBridgeJsonRequest struct {
	Username string `json:"username" example:"john_doe"`
}

type ClientWebhookJsonRequest struct {
	Username   string `json:"username" example:"john_doe"`
	DeviceName string `json:"device_name" example:"wa123456789"`
	URL        string `json:"url" example:"https://example.com"`
	Method     string `json:"method" example:"POST"`
}

// LoginResponse represents the response for successful login
// @Description Response payload for successful login
type LoginResponse struct {
	Username    string `json:"username" example:"john_doe"`
	AccessToken string `json:"access_token" example:"syt_YWxwaGE..."`
	Status      string `json:"status" example:"logged in"`
}

// ErrorResponse represents an error response
// @Description Response payload for error cases
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Details string `json:"details,omitempty" example:"Username must be 3-32 characters"`
}

// MessageResponse represents the response for successful message sending
// @Description Response payload for successful message sending
type MessageResponse struct {
	Contact string `json:"contact" example:"+1234567890"`
	EventID string `json:"event_id" example:"$1234567890abcdef"`
	Message string `json:"message" example:"Hello, world!"`
	Status  string `json:"status" example:"sent"`
}

// DeviceResponse represents the response for successful device addition
// @Description Response payload for successful device addition. The websocket_url is used to establish a connection that:
// @Description - Receives media/images from the platform bridge
// @Description - Handles login synchronization events
// @Description - Receives existing active sessions if available
// @Description - Closes when receiving nil data (indicating end of session or error)
type DeviceResponse struct {
	WebsocketURL string `json:"websocket_url" example:"ws://localhost:8080/ws/telegram/john_doe"`
}

// Webhook represents a webhook configuration
// @Description Represents a webhook structure with device name, URL, method, and timestamp
// @name Webhook
// @type object
type Webhook struct {
	ID             int    `json:"id"`
	ClientUsername string `json:"client_username"`
	DeviceName     string `json:"device_name"`
	URL            string `json:"url"`
	Method         string `json:"method"`
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

	// Remove plus sign if present
	contact = strings.TrimPrefix(contact, "+")

	// E.164 format validation: [country code][number], total length 8-15 digits
	validContact := regexp.MustCompile(`^[1-9]\d{7,14}$`)
	if !validContact.MatchString(contact) {
		return "", fmt.Errorf("contact must be a valid E.164 phone number (e.g., 1234567890 or +1234567890)")
	}

	return contact, nil
}

func sanitizeDeviceName(deviceName string) (string, error) {
	// Remove any whitespace
	deviceName = strings.TrimSpace(deviceName)

	// Device name should be 2-20 characters and contain only letters and numbers
	validDeviceName := regexp.MustCompile(`^[a-z0-9]{2,20}$`)
	if !validDeviceName.MatchString(deviceName) {
		return "", fmt.Errorf("device name must be 2-20 characters and contain only letters and numbers")
	}

	return deviceName, nil
}

// Helper function to extract Bearer token from Authorization header
func extractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is required")
	}

	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("Authorization header must start with 'Bearer '")
	}

	// Extract the token (remove "Bearer " prefix)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("Bearer token cannot be empty")
	}

	return token, nil
}

// ApiLogin godoc
// @Summary Logs a user into the Matrix server
// @Description Authenticates a user and returns an access token
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

	client, err := mautrix.NewClient(homeServer, id.NewUserID(username, cfg.HomeServerDomain), cfg.User.AccessToken)
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	controller := Controller{
		Client:   client,
		Username: username,
		UserID:   client.UserID,
	}
	if err := controller.LoginProcess(password); err != nil {
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
// @Description Registers a new user and returns an access token
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

	controller := Controller{
		Client:   client,
		Username: username,
	}

	if err := controller.CreateProcess(password); err != nil {
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
// @Summary Sends a message to a contact through a platform bridge
// @Description Sends a message to a contact through the specified platform bridge. The message can include text and optional file data.
// @Description The function validates and sanitizes all input fields according to the following rules:
// @Description - Username: 3-32 characters, letters, numbers, and underscores only
// @Description - Message: 1-4096 characters, cannot be empty
// @Description - Device name: 2-20 characters, letters and numbers only
// @Description - Contact: Valid E.164 phone number format (8-15 digits)
// @Description - Platform: 2-20 characters, letters and numbers only
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name (2-20 characters, letters and numbers only)" example:"wa"
// @Param   contact path string true "Contact ID (E.164 phone number without the plus sign, 8-15 digits)" example:"1234567890"
// @Param   Authorization header string true "Bearer token" example:"Bearer syt_YWxwaGE..."
// @Param   payload body ClientMessageJsonRequeset true "Message Payload"
// @Success 200 {object} map[string]interface{} "Message sent successfully" example:{"contact":"1234567890","message":"Hello, world!","status":"sent"}
// @Failure 400 {object} ErrorResponse "Invalid request - validation errors for username, message, device_name, platform, or contact"
// @Failure 401 {object} ErrorResponse "Invalid or missing Bearer token"
// @Failure 500 {object} ErrorResponse "Failed to send message or internal server error"
// @Router /{platform}/message/{contact} [post]
func ApiSendMessage(c *gin.Context) {
	var req ClientMessageJsonRequeset

	// Extract Bearer token from Authorization header
	accessToken, err := extractBearerToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Sanitize platform and contact parameters
	platform, err := sanitizePlatform(c.Param("platform"))
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

	// Validate required fields
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}
	if req.DeviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name is required"})
		return
	}

	// Sanitize username
	username, err := sanitizeUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize message
	message, err := sanitizeMessage(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize device name
	deviceName, err := sanitizeDeviceName(req.DeviceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, _ := (&Conf{}).getConf()
	homeServer := cfg.HomeServer

	client, err := mautrix.NewClient(homeServer, id.NewUserID(username, cfg.HomeServerDomain), accessToken)
	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	// Validate access token
	matrixClient := MatrixClient{
		Client: client,
	}
	_, err = matrixClient.LoadActiveSessionsByAccessToken(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token", "details": err.Error()})
		return
	}

	controller := Controller{
		Client: client,
		UserID: client.UserID,
	}

	err = controller.SendMessage(username, message, contactID, platform, deviceName, req.FileData)

	if err != nil {
		log.Printf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact": contactID,
		"message": message,
		"status":  "sent",
	})
}

// ApiAddDevice godoc
// @Summary Adds a device for a given platform
// @Description Registers a new device connection for the specified platform and establishes a websocket connection.
// @Description The websocket connection will:
// @Description - Receive media/images from the platform bridge
// @Description - Handle login synchronization events
// @Description - Send existing active sessions if available
// @Description - Close connection when receiving nil data (indicating end of session or error)
// @Description Here are various platforms supported:
// @Description 'wa' (for WhatsApp)
// @Description 'signal' (for Signal)
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name" example:"wa"
// @Param   Authorization header string true "Bearer token" example:"Bearer syt_YWxwaGE..."
// @Param   payload body ClientBridgeJsonRequest true "Device Payload"
// @Success 200 {object} DeviceResponse "Successfully added device and established websocket connection"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid or missing Bearer token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /{platform}/devices [post]
func ApiAddDevice(c *gin.Context) {
	var bridgeJsonRequest ClientBridgeJsonRequest

	// Extract Bearer token from Authorization header
	accessToken, err := extractBearerToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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

	client, err := mautrix.NewClient(
		cfg.HomeServer, id.NewUserID(username, cfg.HomeServerDomain), accessToken)

	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	matrixClient := MatrixClient{
		Client: client,
	}
	_, err = matrixClient.LoadActiveSessionsByAccessToken(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token", "details": err.Error()})
		return
	}

	controller := Controller{
		Client: client,
		UserID: client.UserID,
	}

	websocketUrl, err := controller.AddDevice(username, platformName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"websocket_url": websocketUrl,
	})
}

// ApiListDevices godoc
// @Summary Lists devices for a given platform
// @Description Retrieves all active devices for the specified platform and user
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name" example:"wa"
// @Param   Authorization header string true "Bearer token" example:"Bearer syt_YWxwaGE..."
// @Param   payload body ClientBridgeJsonRequest true "Device List Request"
// @Success 200 {object} map[string]interface{} "List of devices"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid or missing Bearer token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /{platform}/list/devices [post]
func ApiListDevices(c *gin.Context) {
	var bridgeJsonRequest ClientBridgeJsonRequest

	// Extract Bearer token from Authorization header
	accessToken, err := extractBearerToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&bridgeJsonRequest); err != nil {
		log.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	username, err := sanitizeUsername(bridgeJsonRequest.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := mautrix.NewClient(
		cfg.HomeServer, id.NewUserID(username, cfg.HomeServerDomain), accessToken)

	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	controller := Controller{
		Client: client,
		UserID: client.UserID,
	}

	platformName, err := sanitizePlatform(c.Param("platform"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("Listing devices for", username, platformName)
	devices, err := controller.ListDevices(username, platformName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

func ApiListWebhooks(c *gin.Context) {
}

// @Summary Adds a webhook for a given device
// @Description Adds a webhook for a given device
// @Accept  json
// @Produce  json
// @Param   platform path string true "Platform Name" example:"wa"
// @Param   device_name path string true "Device Name" example:"wa123456789"
// @Param   url query string true "URL" example:"https://example.com"
// @Param   method query string true "Method" example:"POST"
// @Param   Authorization header string true "Bearer token" example:"Bearer syt_YWxwaGE..."
// @Success 200 {object} map[string]interface{} "Webhook added successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid or missing Bearer token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /{platform}/device/{device_name}/webhook [post]
func ApiAddWebhook(c *gin.Context) {
	var webhookJsonRequest ClientWebhookJsonRequest

	// Extract Bearer token from Authorization header
	accessToken, err := extractBearerToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	username, err := sanitizeUsername(webhookJsonRequest.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := mautrix.NewClient(
		cfg.HomeServer, id.NewUserID(username, cfg.HomeServerDomain), accessToken)

	if err != nil {
		log.Printf("Failed to create Matrix client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not initialize client"})
		return
	}

	controller := Controller{
		Client:   client,
		Username: username,
		UserID:   client.UserID,
	}

	err = controller.AddWebhook(webhookJsonRequest.DeviceName, webhookJsonRequest.URL, webhookJsonRequest.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})

}

func ApiDeleteWebhook(c *gin.Context) {

}

func ApiDeleteDevice(c *gin.Context) {

}

func ApiDeletePlatform(c *gin.Context) {

}

func ApiDeleteAccount(c *gin.Context) {

}

func main() {
	if cfgError != nil {
		panic(cfgError)
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
	router.POST("/:platform/devices", ApiAddDevice)
	router.POST("/:platform/message/:contact", ApiSendMessage)

	router.POST("/:platform/list/devices", ApiListDevices)
	router.POST("/:platform/list/webhooks", ApiListWebhooks)
	router.POST("/:platform/device/:device_name/webhook", ApiAddWebhook)

	router.DELETE("/", ApiDeleteAccount)
	router.DELETE("/devices/:device_id", ApiDeleteDevice)
	router.DELETE("/platforms/:platform/devices/device_id", ApiDeletePlatform)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files for documentation
	router.Static("/_static", "./tutorials/_build/html/_static")

	// router.LoadHTMLFiles("./tutorials/_build/html/index.html")
	router.LoadHTMLGlob("./tutorials/_build/html/*.html")
	router.GET("/tutorials", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	router.GET("/index.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	router.GET("/getting-started.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "getting-started.html", gin.H{})
	})
	router.GET("/adding-devices.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "adding-devices.html", gin.H{})
	})
	router.GET("/listing-devices.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "listing-devices.html", gin.H{})
	})
	router.GET("/sending-messages.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "sending-messages.html", gin.H{})
	})
	router.GET("/user-management.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user-management.html", gin.H{})
	})
	router.GET("/search.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "search.html", gin.H{})
	})
	router.GET("/genindex.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "genindex.html", gin.H{})
	})
	router.GET("/py-modindex.html", func(c *gin.Context) {
		// This file doesn't exist, redirect to genindex instead
		c.Redirect(http.StatusMovedPermanently, "/genindex.html")
	})

	ks.Init()

	host := cfg.Server.Host
	port := cfg.Server.Port

	tlsCert := cfg.Server.Tls.Crt
	tlsKey := cfg.Server.Tls.Key

	go func() {
		err := (&MatrixClient{}).SyncAllClients()
		if err != nil {
			panic(err)
		}
	}()

	if cfg.Websocket.Tls.Crt != "" && cfg.Websocket.Tls.Key != "" {
		go func() {
			err := MainWebsocket(true)
			if err != nil {
				panic(err)
			}
		}()
	} else {
		go func() {
			err := MainWebsocket(false)
			if err != nil {
				panic(err)
			}
		}()
	}

	if tlsCert != "" && tlsKey != "" {
		router.RunTLS(fmt.Sprintf(":%s", port), tlsCert, tlsKey)
	} else {
		router.Run(fmt.Sprintf("%s:%s", host, port))
	}
}
