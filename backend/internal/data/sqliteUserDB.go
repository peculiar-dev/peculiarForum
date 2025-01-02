package data

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type User struct {
	Username  string    // Username of this user
	Created   time.Time // Creation timestamp
	LastLogin time.Time // Last Login timestamp
	Theme     string    // user's current theme

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
		This struct will be a row on a table.

		Username  string    // Username of this user
		Created   time.Time // Creation timestamp
		LastLogin time.Time // Last Login timestamp
		Theme     string    // user's current theme

	*/

	createTableSQL := `CREATE TABLE user (
	"username" TEXT,
	"created_at" DATETIME,
	"lastlogin_at" DATETIME,
	"theme" TEXT
	);`

	log.Println("Create User table...")
	statement, err := db.database.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("user table created")
}

// probably move to a fileUtil package if I need it twice.
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func (db *SqliteUserDB) InsertUser(user *User) {
	/*
		Insert new user into database, give it a default theme (light), and set up default user download directory
		and default user images.
	*/
	currentTime := time.Now()

	log.Println("Inserting user record ...")
	insertUserSQL := `INSERT INTO user(username, created_at, lastlogin_at, theme) VALUES (?, ?, ?, ?)`
	statement, err := db.database.Prepare(insertUserSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(user.Username, currentTime, currentTime, "light")
	if err != nil {
		log.Fatalln(err.Error())
	}

	dirPath := "./downloads/" + user.Username

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create the directory with 0755 permissions
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			//panic(err)
			log.Println("Error creating file directory:", err)
		}
		println("Directory created successfully.")
	} else if err != nil {
		panic(err)
	} else {
		println("Directory already exists.")
	}

	_, err = copy("./static/themes/light/_user_icon.png", "./downloads/"+user.Username+"/_user_icon.png")
	if err != nil {
		log.Println("Error copying file:", err)
	} else {
		log.Println("File copied successfully.")
	}

}

func (db *SqliteUserDB) UpdateUser(user *User) {
	currentTime := time.Now()

	log.Println("Updating user record ...")
	updateUserSQL := `UPDATE user SET theme = ?, lastlogin_at = ?  WHERE username = ?`
	statement, err := db.database.Prepare(updateUserSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(user.Theme, currentTime, user.Username)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (db *SqliteUserDB) LoadTestUsers() {

	//db.InsertUser(&User{Username: "test", Created: time.Now(), LastLogin: time.Now(), Theme: "light"})
	db.InsertUser(&User{Username: "test2", Created: time.Now(), LastLogin: time.Now(), Theme: "light"})
	db.InsertUser(&User{Username: "test3", Created: time.Now(), LastLogin: time.Now(), Theme: "light"})
	db.InsertUser(&User{Username: "test4", Created: time.Now(), LastLogin: time.Now(), Theme: "light"})
	db.InsertUser(&User{Username: "test5", Created: time.Now(), LastLogin: time.Now(), Theme: "light"})

}

func (db *SqliteUserDB) GetUsers() *[]User {

	var users []User
	var username string
	var created time.Time
	var lastLogin time.Time
	var theme string

	rows, err := db.database.Query(`SELECT username, created_at, lastlogin_at, theme
    					   FROM user
						   ORDER BY username;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println("get Users:")
	for rows.Next() {
		rows.Scan(&username, &created, &lastLogin, &theme)
		log.Println("Username: ", username)

		users = append(users, User{Username: username, Created: created, LastLogin: lastLogin, Theme: theme})
	}

	return &users
}

func (db *SqliteUserDB) GetUser(username string) *User {

	var user User
	var created time.Time
	var lastLogin time.Time
	var theme string

	rows, err := db.database.Query(`SELECT username, created_at, lastlogin_at, theme
    					   FROM user
						   Where username = ?
						   ORDER BY username;`, username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println("got User:" + username)
	for rows.Next() {
		rows.Scan(&username, &created, &lastLogin, &theme)
		user = User{Username: username, Created: created, LastLogin: lastLogin, Theme: theme}
	}

	return &user
}
