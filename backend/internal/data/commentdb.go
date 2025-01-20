package data

import "database/sql"

//These data objects should give the main the option to create a separate database for a per-table structure, or
//allow the system to share a single instance of a database, by taking a pointer to an existing database.

type Commentdb interface {
	Getdb() *sql.DB
	Setdb(*sql.DB)
	InitDB(initialize, debug bool)
	LoadTestComments()
	GetRootMail(username string) *[]Comment
	GetRootComments(username string) *[]Comment
	GetCommentsFromTo(username string, startIdx, endIdx int) *[]Comment
	GetMailComments(parentID string, username string) *[]Comment
	GetChildComments(parentID string, username string) *[]Comment
	CreateCommentTable()
	InsertComment(id, rootID, user, message, parent string, root bool, sticky bool)
	EditComment(id, message, parent string, root bool, sticky bool)
	EditCommentPic(id, picture string)
}
