package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var roomTypes = NewRoomTypesRegistry()

func NewRoomTypesRegistry() *RoomTypes {
	management := RoomType{0}

	return &RoomTypes{
		Management: management,
		types:      []*RoomType{&management},
	}
}

func (r *RoomType) Parse() int {
	for _, roomType := range roomTypes.types {
		if roomType == r {
			return roomType.IntValue
		}
	}
	return -1
}

type RoomTypes struct {
	Management RoomType

	types []*RoomType
}

type RoomType struct {
	IntValue int
}

type Rooms struct {
	ID       id.RoomID
	Channel  chan *event.Event
	Type     RoomType
	isBridge bool
	User     Users
	Members  map[string]string
}

func (r *Rooms) ListenJoinedRooms(
	client *mautrix.Client,
	callback IncomingMessageCallback,
) {
	log.Println(">> Begin listening...", client.UserID.String())
	joinedRooms, err := r.JoinedRooms(client)
	log.Printf("[+] Joined rooms: %v\n", joinedRooms)

	if err != nil {
		log.Fatalf("[-] Failed to fetch rooms: %v", err)
	}

	// map RoomID -> room channel
	var roomChannels = make(map[id.RoomID]chan *event.Event)

	chanBufferSize := 500 //TODO: move to a config file

	var wg sync.WaitGroup
	for _, roomId := range joinedRooms {
		channel := make(chan *event.Event, chanBufferSize)
		roomChannels[roomId] = channel

		room := Rooms{
			ID:      roomId,
			Channel: channel,
			User:    r.User,
		}

		wg.Add(1)
		go func(rm Rooms) {
			defer wg.Done()
			rm.GetRoomMessages(client, callback)
		}(room)
	}

	go func() {
		for evt := range r.Channel {
			// evt := <-r.Channel
			// log.Println("[*] Dispatching event to room:", evt.RoomID)

			ch, exists := roomChannels[evt.RoomID]
			if !exists {
				// New room? Start listening dynamically
				channel := make(chan *event.Event, chanBufferSize)
				roomChannels[evt.RoomID] = channel

				newRoom := Rooms{
					ID:      evt.RoomID,
					Channel: channel,
					User:    r.User,
				}

				wg.Add(1)
				go func(rm Rooms) {
					defer wg.Done()
					rm.GetRoomMessages(client, callback)
				}(newRoom)
				ch = channel
			}

			// Send event (non-blocking safe send)
			select {
			case ch <- evt:
			default:
				log.Printf("[-] Dropping event: channel for %v is full", evt.RoomID)
			}
		}
	}()

	wg.Wait()
	log.Println("[-] Finished listening to rooms...")
}

func (r *Rooms) JoinedRooms(
	client *mautrix.Client,
) ([]id.RoomID, error) {
	resp, err := client.JoinedRooms(context.Background())

	if err != nil {
		log.Fatalf("Failed to fetche rooms: %v", err)
	}

	return resp.JoinedRooms, err
}

type IncomingMessageMetaData struct {
	DisplayName string
	MxID        id.UserID
	Type        string
	Message     MessageMetaData
	RoomID      id.RoomID
}

type MessageMetaData struct {
	Content   event.Content
	Timestamp int64
	Type      interface{}
}

func (r *Rooms) SendRoomMessages(client *mautrix.Client, message string) (*mautrix.RespSendEvent, error) {
	log.Printf("[+] Sending message: %s to %v - %s\n", message, r.ID, client.AccessToken)

	resp, err := client.SendText(
		context.Background(),
		id.RoomID(r.ID),
		message,
	)

	return resp, err
}

type IncomingMessageCallback func(IncomingMessageMetaData, error)

var MessageTypeSending = "sending"
var MessageTypeReceiving = "receiving"

func (r *Rooms) GetRoomMessages(
	client *mautrix.Client,
	callback IncomingMessageCallback,
) {
	log.Println("[+] Getting messages for: ", r.ID)
	for evt := range r.Channel {
		if evt.Type == event.EventMessage && r.ID == evt.RoomID {

			userProfile, _ := client.GetProfile(context.Background(), evt.Sender)

			if isHandled, _ := r.HandleMessage(evt); isHandled {
				return
			}

			_type := MessageTypeReceiving
			if evt.Sender == client.UserID {
				_type = MessageTypeSending
			}

			if userProfile != nil {
				incomingMessageMetaData := IncomingMessageMetaData{
					DisplayName: userProfile.DisplayName,
					MxID:        evt.Sender,
					Type:        _type,
					Message: MessageMetaData{
						Content:   evt.Content,
						Timestamp: evt.Timestamp,
						Type:      evt.Content.Raw["msgtype"],
					},
				}

				go callback(incomingMessageMetaData, nil)
			}
		}
	}
}

func (r *Rooms) CreateRoom(
	client *mautrix.Client,
	name string,
	members string,
	_type RoomType,
	isBridge bool,
) (id.RoomID, error) {
	resp, err := client.CreateRoom(context.Background(), &mautrix.ReqCreateRoom{
		Invite:   []id.UserID{id.UserID(members)},
		IsDirect: true,
		// Preset:     "private_chat",
		Preset:     "trusted_private_chat",
		Visibility: "private",
	})

	if err != nil {
		return "", err
	}

	r.ID = resp.RoomID
	r.Type = _type

	var clientDB ClientDB = ClientDB{
		username: r.User.name,
		filepath: "db/" + r.User.name + ".db",
	}
	fmt.Println(clientDB)

	clientDB.Init()
	if err := clientDB.StoreRooms(
		r.ID.String(),
		name,
		members,
		_type.Parse(),
		isBridge,
	); err != nil {
		return resp.RoomID, err
	}

	return resp.RoomID, nil
}

func (r *Rooms) Join(
	client *mautrix.Client,
	roomId id.RoomID,
) error {
	log.Println("[*] Joining room:", roomId)
	_, err := client.JoinRoomByID(context.Background(), roomId)
	return err
}

func ParseImage(client *mautrix.Client, url string) ([]byte, error) {
	fmt.Printf(">>\tParsing image for: %v\n", url)
	contentUrl, err := id.ParseContentURI(url)
	if err != nil {
		panic(err)
	}
	return client.DownloadBytes(context.Background(), contentUrl)
}

func (r *Rooms) HandleMessage(evt *event.Event) (bool, error) {
	// check room
	// check template

	if evt.Type == event.EventMessage {
		var clientDB ClientDB = ClientDB{
			username: r.User.name,
			filepath: "db/" + r.User.name + ".db",
		}

		clientDB.Init()
		defer clientDB.Close()

		room, err := clientDB.FetchRooms(evt.RoomID.String())

		if err != nil {
			return false, err
		}

		if !room.isBridge {
			return false, nil
		}

		// log.Println("[+] BRIDGE| New message:", evt.Content.AsMessage().Body)
		// log.Println(evt.Content.Raw)
		log.Printf("Bridge message - %v", evt.Content.Raw)
		return true, nil
	}
	return false, nil
}
