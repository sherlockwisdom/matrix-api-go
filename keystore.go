package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Keystore struct {
	connection *sql.DB
	filepath   string
}

func (ks *Keystore) Init() {
	db, err := sql.Open("sqlite3", ks.filepath)
	if err != nil {
		panic(err)
	}
	ks.connection = db

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users ( 
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		username TEXT NOT NULL, 
		accessToken TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)

	if err != nil {
		panic(err)
	}
}

func (ks *Keystore) CreateUser(username string, accessToken string) error {
	tx, err := ks.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO users (username, accessToken) values(?,?)`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(username, accessToken)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ks *Keystore) FetchUser(username string) (Users, error) {
	stmt, err := ks.connection.Prepare("select id, username, accessToken from users where username = ?")
	if err != nil {
		return Users{}, err
	}

	defer stmt.Close()

	var id int
	var _username string
	var _accessToken string
	err = stmt.QueryRow(username).Scan(&id, &_username, &_accessToken)
	if err != nil {
		return Users{}, err
	}

	return Users{ID: id, Username: _username, AccessToken: _accessToken}, nil
}

func (ks *Keystore) FetchAllUsers() ([]Users, error) {
	stmt, err := ks.connection.Prepare("select id, username, accessToken from users")
	if err != nil {
		return []Users{}, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return []Users{}, err
	}

	defer rows.Close()

	var users []Users
	for rows.Next() {
		var id int
		var _username string
		var _accessToken string

		err = rows.Scan(&id, &_username, &_accessToken)
		if err != nil {
			return []Users{}, err
		}

		users = append(users, Users{ID: id, Username: _username, AccessToken: _accessToken})
	}

	return users, nil
}

// https://github.com/mattn/go-sqlite3/blob/v1.14.28/_example/simple/simple.go
func (clientDb *ClientDB) Init() error {
	db, err := sql.Open("sqlite3", clientDb.filepath)
	if err != nil {
		return err
	}

	clientDb.connection = db

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS clients ( 
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	username TEXT NOT NULL, 
	password TEXT NOT NULL,
	accessToken TEXT NOT NULL, 
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS rooms ( 
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	clientUsername TEXT NOT NULL,
	roomID TEXT NOT NULL,
	name TEXT NOT NULL,
	members TEXT NOT NULL,
	type INTEGER NOT NULL,
	isBridge INTEGER NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)

	if err != nil {
		return err
	}
	return err
}

func (clientDb *ClientDB) Authenticate(username string, password string) (bool, error) {
	query := `SELECT COUNT(*) FROM clients WHERE username = ? AND password = ?`

	var count int
	err := clientDb.connection.QueryRow(query, username, password).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("authentication query failed: %w", err)
	}

	if count == 0 {
		log.Printf("[-] Authentication failed for user: %s", username)
		return false, nil
	}

	log.Printf("[+] Authentication successful for user: %s", username)
	return true, nil
}

func (clientDb *ClientDB) Store(accessToken string, password string) error {
	tx, err := clientDb.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO clients (username, accessToken, password) values(?,?,?)`)
	if err != nil {
		return err
	}

	defer stmt.Close()
	log.Println("[+] Storing for username:", clientDb.username, ", AT:", accessToken)

	_, err = stmt.Exec(clientDb.username, accessToken, password)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (clientDb *ClientDB) Fetch() (string, error) {
	fmt.Println("- Fetching username:", clientDb.filepath)
	stmt, err := clientDb.connection.Prepare("select accessToken from clients where username = ?")
	var accessToken string

	if err != nil {
		return accessToken, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(clientDb.username).Scan(&accessToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return accessToken, nil
		}
		return accessToken, err
	}
	return accessToken, err
}

func (clientDb *ClientDB) Close() {
	defer clientDb.connection.Close()
}

func (clientDb *ClientDB) StoreRooms(
	roomID string,
	platformName string,
	members string,
	_type int,
	isBridge bool,
) error {
	log.Println("[+] Storing to rooms:", roomID, platformName, members, _type, isBridge)
	tx, err := clientDb.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		`INSERT INTO rooms (clientUsername, roomID, name, members, type, isBridge) values(?,?,?,?,?,?)`,
	)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(clientDb.username, roomID, platformName, members, _type, isBridge)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (clientDb *ClientDB) FetchRooms(roomID string) (Rooms, error) {
	stmt, err := clientDb.connection.Prepare(
		"select clientUsername, roomID, name, members, type, isBridge from rooms where roomID = ?",
	)
	if err != nil {
		return Rooms{}, err
	}
	var clientUsername string
	var _roomID string
	var name string
	var members string
	var _type int
	var isBridge bool

	defer stmt.Close()

	err = stmt.QueryRow(roomID).Scan(&clientUsername, &_roomID, &name, &members, &_type, &isBridge)
	if err != nil {
		if err == sql.ErrNoRows {
			return Rooms{}, nil
		}
		return Rooms{}, err
	}

	var room = Rooms{
		ID:       id.RoomID(_roomID),
		Channel:  make(chan *event.Event),
		Type:     RoomType(_type),
		isBridge: isBridge,
		Members: map[string]string{
			name: members,
		},
	}

	return room, err
}

func (clientDb *ClientDB) FetchRoomsByMembers(name string) (Rooms, error) {
	log.Println("Fetching room members for", name, clientDb.filepath)
	stmt, err := clientDb.connection.Prepare(
		"select clientUsername, roomID, name, members, type, isBridge from rooms where name = ?",
	)
	if err != nil {
		return Rooms{}, err
	}
	var clientUsername string
	var _roomID string
	var _name string
	var members string
	var _type int
	var isBridge bool

	defer stmt.Close()

	err = stmt.QueryRow(name).Scan(&clientUsername, &_roomID, &_name, &members, &_type, &isBridge)
	if err != nil {
		if err == sql.ErrNoRows {
			return Rooms{}, nil
		}
		return Rooms{}, err
	}

	var room = Rooms{
		ID:       id.RoomID(_roomID),
		Channel:  make(chan *event.Event),
		Type:     RoomType(_type),
		isBridge: isBridge,
		Members: map[string]string{
			_name: members,
		},
	}

	return room, err
}
