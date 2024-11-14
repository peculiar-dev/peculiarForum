package commentdb

import "database/sql"

type Commentdb interface {
	Getdb() *sql.DB
	InitDB()
	LoadTestComments()
	GetRootMail(username string) *[]Comment
	GetRootComments(username string) *[]Comment
	GetMailComments(parentID string, username string) *[]Comment
	GetChildComments(parentID string, username string) *[]Comment
	CreateCommentTable()
	InsertComment(id, user, message, parent string, root bool, sticky bool)
	EditComment(id, message, parent string, root bool, sticky bool)
	EditCommentPic(id, picture string)
}
