package data

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Notification struct {
	Sender      string    // the user who sent this notification
	Reciever    string    // the user who is to recieve this notification
	Created     time.Time // Creation timestamp
	CommentUUID string    // the UUID of the relevant comment

}

type SqliteNotificationDB struct {
	database *sql.DB
}

func (db *SqliteNotificationDB) Setdb(newdb *sql.DB) {
	db.database = newdb
}

func (db *SqliteNotificationDB) Getdb() *sql.DB {
	return db.database
}

func NewSqliteNotificationDB() *SqliteNotificationDB {
	return &SqliteNotificationDB{}
}

func (db *SqliteNotificationDB) InitDB() {
	os.Remove("sqlite-notification-database.db") // I delete the file to avoid duplicated records.
	// SQLite is a file based database.

	log.Println("Creating sqlite-notification-database.db...")
	file, err := os.Create("sqlite-notification-database.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-notification-database.db created")

	sqliteDatabase, error := sql.Open("sqlite3", "./sqlite-notification-database.db") // Open the created SQLite File
	if error != nil {
		log.Fatal(error.Error())
	}

	db.database = sqliteDatabase

}
