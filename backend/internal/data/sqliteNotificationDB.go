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

func (db *SqliteNotificationDB) CreateNotificationTable() {
	/*
		This struct will be a row on a table.

		Sender      string    // the user who sent this notification
		Reciever    string    // the user who is to recieve this notification
		Created     time.Time // Creation timestamp
		CommentUUID string    // the UUID of the relevant comment

	*/

	createTableSQL := `CREATE TABLE notification (
	"sender" TEXT,
	"reciever" TEXT,
	"commentUUID" TEXT,
	"created_at" DATETIME
	);`

	log.Println("Create Notification table...")
	statement, err := db.database.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("Notification table created")
}

func (db *SqliteNotificationDB) InsertNotification(notification Notification) {
	currentTime := time.Now()

	log.Println("Inserting notification record ...")
	insertNotificationSQL := `INSERT INTO notification(sender, reciever, commentUUID, created_at) VALUES (?, ?, ?, ?)`
	statement, err := db.database.Prepare(insertNotificationSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(notification.Sender, notification.Reciever, notification.CommentUUID, currentTime)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (db *SqliteNotificationDB) LoadTestNotifications() {

	db.InsertNotification(Notification{Sender: "test2", Reciever: "test", CommentUUID: "id-1", Created: time.Now()})
	db.InsertNotification(Notification{Sender: "test2", Reciever: "test", CommentUUID: "id-1", Created: time.Now()})
	db.InsertNotification(Notification{Sender: "test", Reciever: "test2", CommentUUID: "id-5", Created: time.Now()})
	db.InsertNotification(Notification{Sender: "test", Reciever: "test2", CommentUUID: "id-5", Created: time.Now()})

}

func (db *SqliteNotificationDB) GetNotifications(username string) *[]Notification {

	var notifications []Notification
	var sender string
	var commentUUID string
	var created time.Time

	rows, err := db.database.Query(`SELECT sender, commentUUID, created_at
    					   FROM notification
						   WHERE reciever = ?
						   ORDER BY created_at;`, username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println(username + " notifications:")
	for rows.Next() {
		rows.Scan(&sender, &commentUUID, &created)

		notifications = append(notifications, Notification{Sender: sender, Reciever: username, CommentUUID: commentUUID, Created: created})
	}

	return &notifications
}
