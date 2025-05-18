package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// https://github.com/mattn/go-sqlite3/blob/v1.14.28/_example/simple/simple.go

type ClientDBInterface interface {
	Init() error
	Close()
}

type ClientDB struct {
	connection *sql.DB
	username   string
	filepath   string
}

func (clientDb *ClientDB) Init() error {
	db, err := sql.Open("sqlite3", clientDb.filepath)
	if err != nil {
		return err
	}

	clientDb.connection = db

	_, err = db.Exec(`
	CREATE CLIENTS IF NOT EXISTS clients ( 
	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	username TEXT NOT NULL, 
	accessToken BLOB NOT NULL, 
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	// CREATE CLIENT_BRIDGES IF NOT EXISTS clients ( 
	// id INTEGER PRIMARY KEY AUTOINCREMENT, 
	// username TEXT NOT NULL, 
	// accessToken BLOB NOT NULL, 
	// timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	// );
	`)

	if err != nil {
		return err
	}
	return err
}

func (clientDb *ClientDB) Store(accessToken []byte) error {
	tx, err := clientDb.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO clients (username, accessToken) values(?,?)`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(clientDb.username, accessToken)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (clientDb *ClientDB) Fetch() ([]byte, error) {
	stmt, err := clientDb.connection.Prepare("select accessToken from clients where username = ?")
	var accessToken []byte

	if err != nil {
		return accessToken, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(clientDb.username).Scan(&accessToken)
	if err != nil {
		return accessToken, err
	}
	return accessToken, err
}

func (clientDb *ClientDB) Close() {
	defer clientDb.connection.Close()
}
