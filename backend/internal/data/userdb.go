package data

import "database/sql"

//These data objects should give the main the option to create a separate database for a per-table structure, or
//allow the system to share a single instance of a database, by taking a pointer to an existing database.

type Userdb interface {
	Getdb() *sql.DB
	Setdb(*sql.DB)
	InitDB()
	CreateUserTable()
	LoadTestUsers()
	InsertUser(user *User)
	UpdateUser(user *User)
	GetUsers() *[]User
	GetUser(username string) *User
}
