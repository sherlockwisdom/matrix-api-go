package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type ClientJsonRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ApiCreate(c *gin.Context) {
	var clientJsonRequest ClientJsonRequest

	if err := c.BindJSON(&clientJsonRequest); err != nil {
		return
	}

	homeServer := "https://relaysms.me"

	client, err := mautrix.NewClient(homeServer, "", "")
	if err != nil {
		log.Fatalln(err)
		return
	}

	var bridge Bridges
	var room = Rooms{
		Channel: make(chan *event.Event),
		Bridge:  bridge,
	}

	if err := CreateProcess(
		client,
		&room,
		clientJsonRequest.Username,
		clientJsonRequest.Password,
	); err != nil {
		log.Fatalln(err)
		return
	}

	c.IndentedJSON(http.StatusCreated, clientJsonRequest)
}

func main() {
	router := gin.Default()
	router.POST("/create", ApiCreate)

	router.Run("localhost:8080")
}
