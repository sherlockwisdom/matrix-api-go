package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
	Bridge   Bridges
	Type     RoomType
	isBridge bool
}

func (r *Rooms) ListenJoinedRooms(
	client *mautrix.Client,
) {
	log.Println(">> Begin listening...", client.UserID.String())
	joinedRooms, err := r.JoinedRooms(client)
	log.Printf("[+] Joined rooms: %v\n", joinedRooms)

	if err != nil {
		log.Fatalf("[-] Failed to fetch rooms: %v", err)
	}

	var wg sync.WaitGroup
	for _, roomId := range joinedRooms {
		wg.Add(1)
		var room = Rooms{
			ID:      id.RoomID(roomId),
			Channel: r.Channel,
			Bridge:  Bridges{r.Bridge.username},
		}
		go func() {
			room.GetRoomMessages(client)
			defer wg.Done()
		}()
		go func() {
			room.GetInvites(client)
			defer wg.Done()
		}()
	}
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

func (r *Rooms) GetRoomMessages(
	client *mautrix.Client,
) {
	log.Println("[+] Getting messages for: ", r.ID)
	for {
		evt := <-r.Channel
		if evt.Type == event.EventMessage {
			log.Printf("[*] Room channel parsing %v, %v\n", evt.Sender, evt.RoomID)
			content := evt.Content.Raw

			if isHandled, _ := r.Bridge.HandleMessage(evt); isHandled {
				return
			}

			switch content["msgtype"] {
			case "m.text":
				log.Println("[+] MSG:", content["body"].(string))
			case "m.image":
				log.Println("[+] saving image", evt.Content.Raw)
				rawImage, err := ParseImage(client, content["url"].(string))
				if err != nil {
					panic(err)
				}

				filename := content["filename"]
				if filename == nil {
					filename = content["body"]
				}
				imageDownloadFilepath := "downloads/rooms/" + filename.(string)
				os.WriteFile(imageDownloadFilepath, rawImage, 0644)
				log.Printf("[+] Saved image to room dir: %s\n", imageDownloadFilepath)
			default:
				log.Printf("[-] Type not yet implemented: %v\n", content["msgtype"])
			}
		}
	}
}

func (r *Rooms) CreateRoom(
	client *mautrix.Client,
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
		username: r.Bridge.username,
		filepath: "db/" + r.Bridge.username + ".db",
	}
	fmt.Println(clientDB)

	clientDB.Init()
	if err := clientDB.StoreRooms(
		r.ID.String(),
		members,
		_type.Parse(),
		isBridge,
	); err != nil {
		panic(err)
	}

	return resp.RoomID, nil
}

func (r *Rooms) GetInvites(
	client *mautrix.Client,
) {
	log.Println("[+] Getting invites for: ", r.ID)
	for {
		evt := <-r.Channel

		if evt.Content.AsMember().Membership == event.MembershipInvite {
			if evt.StateKey != nil && *evt.StateKey == client.UserID.String() {
				log.Printf("[+] >> New invite to room: %s from %s\n", evt.RoomID, evt.Sender)
				err := r.Join(client, evt.RoomID)
				if err != nil {
					panic(err)
				}
			}
		}
	}
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
