package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Comment struct {
	Id       string    // the UUID of this comment
	RootId   string    // the UUID of this comment's root comment
	User     string    // The Username of the creator
	Message  string    // The comment message
	Picture  string    // The picture uploaded with this comment
	Root     bool      // Is this a root comment?
	Sticky   bool      // Is this a 'sticky' or 'pinned' comment?
	Editable bool      // not saved in database
	Created  time.Time // Creation timestamp
	Sublist  []Comment // Children of this comment
}

type SqliteCommentDB struct {
	database *sql.DB
}

func (db *SqliteCommentDB) Setdb(newdb *sql.DB) {
	db.database = newdb
}

func (db *SqliteCommentDB) Getdb() *sql.DB {
	return db.database
}

func NewSqliteCommentDB() *SqliteCommentDB {
	return &SqliteCommentDB{}
}

func (db *SqliteCommentDB) InitDB() {
	os.Remove("sqlite-database.db") // I delete the file to avoid duplicated records.
	// SQLite is a file based database.

	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")

	sqliteDatabase, error := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	if error != nil {
		log.Fatal(error.Error())
	}
	//defer sqliteDatabase.Close() // Defer Closing the database

	db.database = sqliteDatabase

	db.CreateCommentTable()
	db.LoadTestComments()

	// DISPLAY INSERTED RECORDS
	//displayComments(sqliteDatabase, &comments)
}

func (db *SqliteCommentDB) LoadTestComments() {

	db.InsertComment("id-1", "", "test", "test message 1", "root", true, false)
	db.InsertComment("id-2", "id-1", "test", "test message 2", "id-1", false, false)
	db.InsertComment("id-3", "id-1", "test", "test message 3", "id-1", false, false)
	db.InsertComment("id-4", "id-1", "test2", "test message 4", "id-3", false, false)
	db.InsertComment("id-5", "", "test2", "test message 5", "root", true, false)
	db.InsertComment("id-6", "id-5", "test", "test message 6", "id-5", false, false)
	db.InsertComment("id-7", "id-5", "test2", "test message 7", "id-5", false, false)
	db.InsertComment("id-8", "", "test", "test mail message 1", "test2-test", true, false)
	db.InsertComment("id-9", "", "test2", "test mail message 2", "test-test2", true, false)

	//test child comment logic.
	db.GetChildComments("id-1", "test")
}

func (db *SqliteCommentDB) GetRootMail(username string) *[]Comment {
	var comments []Comment

	var id string
	var user string
	var message string
	var picture string
	var parent string
	var root bool
	var sticky bool
	var editable bool
	var created time.Time
	//rows, err := db.Query("SELECT * FROM comment where parent = 'root' ")
	rows, err := db.database.Query(`SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE root = ? AND parent LIKE ?
						   ORDER BY created_at DESC;`, 1, "%"+username+"%")

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println("root mail:")
	for rows.Next() {
		rows.Scan(&id, &user, &message, &picture, &parent, &root, &sticky, &created)
		log.Println("Comment ID:", id, " Message:", message, "Parent", parent)
		editable = (username == user)
		comments = append(comments, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
	}

	return &comments

}

func (db *SqliteCommentDB) GetRootComments(username string) *[]Comment {
	var comments []Comment

	var id string
	var user string
	var message string
	var picture string
	var parent string
	var root bool
	var sticky bool
	var editable bool
	var created time.Time
	//rows, err := db.Query("SELECT * FROM comment where parent = 'root' ")
	rows, err := db.database.Query(`SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE root = 1 and parent = 'root'
						   ORDER BY created_at DESC;`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println("root comments:")
	for rows.Next() {
		rows.Scan(&id, &user, &message, &picture, &parent, &root, &sticky, &created)
		log.Println("Comment ID:", id, " Message:", message, "Parent", parent)
		editable = (username == user)
		comments = append(comments, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
	}

	return &comments

}

func (db *SqliteCommentDB) GetMailComments(parentID string, username string) *[]Comment {
	var comments []Comment

	var id string
	var user string
	var message string
	var picture string
	var parent string
	var root bool
	var sticky bool
	var editable bool
	var created time.Time
	//rows, err := db.Query("SELECT * FROM comment where parent = 'root' ")
	rows, err := db.database.Query(`WITH RECURSIVE descendants AS (
    					   SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE id = ?
    
    					   UNION ALL
    
    					   SELECT m.id, m.user, m.message, m.picture, m.parent, m.root, m.sticky, m.created_at
    					   FROM comment m
    					   INNER JOIN descendants d ON m.parent = d.id
						   )
						   SELECT * FROM descendants
						   ORDER BY created_at;`, parentID)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &user, &message, &picture, &parent, &root, &sticky, &created)
		log.Println("Comment ID:", id, " Message:", message, "Parent", parent)
		editable = (username == user)
		comments = append(comments, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
	}

	return &comments

}

func (db *SqliteCommentDB) GetChildComments(parentID string, username string) *[]Comment {
	var comments []Comment

	var id string
	var user string
	var message string
	var picture string
	var parent string
	var root bool
	var sticky bool
	var editable bool
	var created time.Time

	//rows, err := db.Query("SELECT * FROM comment where parent = 'root' ")
	rows, err := db.database.Query(`WITH RECURSIVE descendants AS (
    					   SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE id = ?
    
    					   UNION ALL
    
    					   SELECT m.id, m.user, m.message, m.picture, m.parent, m.root, m.sticky, m.created_at
    					   FROM comment m
    					   INNER JOIN descendants d ON m.parent = d.id
						   )
						   SELECT * FROM descendants
						   ORDER BY created_at;`, parentID)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("In getChildComments")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &user, &message, &picture, &parent, &root, &sticky, &created)
		editable = (username == user)
		log.Println("Comment ID:", id, " Message:", message, " Parent", parent, " Editable:", editable)
		if parent == parentID || root {
			comments = append(comments, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
		} else {
			log.Println("adding ", id, " to parent ", parent)
			addCommentToSublist(&comments, parent, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
		}
	}

	return &comments

}

func (db *SqliteCommentDB) CreateCommentTable() {
	/*
		This struct will be a row on a table. The sublist relationship
		will be maintained by loading all comments with the parent
		of the current comment.

		Id      string
		User    string
		Message string
		Sublist []Comment
	*/

	createTableSQL := `CREATE TABLE comment (
	"id" TEXT,
	"root_id" TEXT,
	"user" TEXT,
	"message" TEXT,
	"picture" TEXT,
	"parent" TEXT,
	"root" BIT,
	"sticky" BIT,
	"created_at" DATETIME
	);`

	log.Println("Create comment table...")
	statement, err := db.database.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("comment table created")
}

func (db *SqliteCommentDB) InsertComment(id, rootID, user, message, parent string, root bool, sticky bool) {
	currentTime := time.Now()

	message = strings.Replace(message, "\n", "<br>", -1)

	log.Println("Inserting comment record ...")
	insertCommentSQL := `INSERT INTO comment(id, root_id, user, message,picture, parent,root,sticky,created_at) VALUES (?,?, ?, ?, ?, ?, ?, ?,?)`
	statement, err := db.database.Prepare(insertCommentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, rootID, user, message, "", parent, root, sticky, currentTime)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (db *SqliteCommentDB) EditComment(id, message, parent string, root bool, sticky bool) {
	//currentTime := time.Now()

	message = strings.Replace(message, "\n", "<br>", -1)

	log.Println("Editing comment record ...")
	//insertCommentSQL := `INSERT INTO comment(id, user, message, parent,root,sticky,created_at) VALUES (?, ?, ?, ?, ?, ?,?)`

	editCommentSQL := `UPDATE comment SET message = ? WHERE id =?`
	statement, err := db.database.Prepare(editCommentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	//_, err = statement.Exec(id, user, message, parent, root, sticky, currentTime)
	_, err = statement.Exec(message, id)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func (db *SqliteCommentDB) EditCommentPic(id, picture string) {
	//currentTime := time.Now()

	log.Println("Editing comment record to change picture ...")
	//insertCommentSQL := `INSERT INTO comment(id, user, message, parent,root,sticky,created_at) VALUES (?, ?, ?, ?, ?, ?,?)`

	editCommentSQL := `UPDATE comment SET picture = ? WHERE id =?`
	statement, err := db.database.Prepare(editCommentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	//_, err = statement.Exec(id, user, message, parent, root, sticky, currentTime)
	_, err = statement.Exec(picture, id)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// AddToSublist searches for an item by its name in a potentially nested structure
func addCommentToSublist(comments *[]Comment, id string, newComment Comment) bool {
	for i, comment := range *comments {
		if comment.Id == id {
			// If found, add newItem to the sublist
			(*comments)[i].Sublist = append((*comments)[i].Sublist, newComment)
			return true
		}
		// Recursive call to search in sublists
		if addCommentToSublist(&comment.Sublist, id, newComment) {
			return true
		}
	}
	return false
}

// Helper function to print comments and their sublists
func PrintComments(comments *[]Comment, indent string) {
	for _, comment := range *comments {
		fmt.Println(indent+"Comment:", comment.Message)
		if len(comment.Sublist) > 0 {
			PrintComments(&comment.Sublist, indent+"  ") // Increase indentation for sublists
		}
	}

}
