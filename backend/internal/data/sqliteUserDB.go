package data

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type User struct {
	Username  string    // Username of this user
	Created   time.Time // Creation timestamp
	LastLogin time.Time // Last Login timestamp

}

type SqliteUserDB struct {
	database *sql.DB
}

func (db *SqliteUserDB) Setdb(newdb *sql.DB) {
	db.database = newdb
}

func (db *SqliteUserDB) Getdb() *sql.DB {
	return db.database
}

func NewSqliteUserDB() *SqliteUserDB {
	return &SqliteUserDB{}
}

func (db *SqliteUserDB) InitDB() {
	os.Remove("sqlite-user-database.db") // I delete the file to avoid duplicated records.
	// SQLite is a file based database.

	log.Println("Creating sqlite-user-database.db...")
	file, err := os.Create("sqlite-user-database.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-user-database.db created")

	sqliteDatabase, error := sql.Open("sqlite3", "./sqlite-user-database.db") // Open the created SQLite File
	if error != nil {
		log.Fatal(error.Error())
	}

	db.database = sqliteDatabase

}

func (db *SqliteUserDB) CreateUserTable() {
	/*
		This struct will be a row on a table. The sublist relationship
		will be maintained by loading all comments with the parent
		of the current comment.

		Username  string    // Username of this user
		Created   time.Time // Creation timestamp
		LastLogin time.Time // Last Login timestamp

	*/

	createTableSQL := `CREATE TABLE user (
	"username" TEXT,
	"created_at" DATETIME,
	"lastlogin_at" DATETIME
	);`

	log.Println("Create User table...")
	statement, err := db.database.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("user table created")
}

func (db *SqliteUserDB) InsertUser(user User) {
	currentTime := time.Now()

	log.Println("Inserting user record ...")
	insertCommentSQL := `INSERT INTO user(username, created_at, lastlogin_at) VALUES (?, ?, ?)`
	statement, err := db.database.Prepare(insertCommentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(user.Username, currentTime, currentTime)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (db *SqliteUserDB) LoadTestComments() {

	db.InsertUser(User{Username: "test", Created: time.Now(), LastLogin: time.Now()})
	db.InsertUser(User{Username: "test2", Created: time.Now(), LastLogin: time.Now()})

}

func (db *SqliteUserDB) GetUsers() *[]User {

	var users []User
	var username string
	var created time.Time
	var lastLogin time.Time

	rows, err := db.database.Query(`SELECT username, created_at, lastlogin_at
    					   FROM user
						   ORDER BY username;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println("root comments:")
	for rows.Next() {
		rows.Scan(&username, &created, lastLogin)
		log.Println("Username: ", username)

		users = append(users, User{Username: username, Created: created, LastLogin: lastLogin})
	}

	return &users
}
