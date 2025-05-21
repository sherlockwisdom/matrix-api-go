package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

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
		panic(err)
	}
	return accessToken, err
}

func (clientDb *ClientDB) Close() {
	defer clientDb.connection.Close()
}

func (clientDb *ClientDB) StoreRooms(
	roomID string,
	members string,
	_type int,
	isBridge bool,
) error {
	log.Println("[+] Storing to rooms:", roomID)
	tx, err := clientDb.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		`INSERT INTO rooms (clientUsername, roomID, members, type, isBridge) values(?,?,?,?,?)`,
	)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(clientDb.username, roomID, members, _type, isBridge)
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
		"select clientUsername, roomID, members, type, isBridge from rooms where roomID = ?",
	)
	if err != nil {
		return Rooms{}, err
	}
	var clientUsername string
	var _roomID string
	var members string
	var _type int
	var isBridge bool

	defer stmt.Close()

	fmt.Println(stmt)

	err = stmt.QueryRow(roomID).Scan(&clientUsername, &_roomID, &members, &_type, &isBridge)
	if err != nil {
		panic(err)
	}

	var room = Rooms{
		ID:       id.RoomID(_roomID),
		Channel:  make(chan *event.Event),
		Type:     RoomType{_type},
		isBridge: isBridge,
	}

	return room, err
}
