package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Comment struct {
	Id       string
	User     string
	Message  string
	Picture  string
	Root     bool
	Sticky   bool
	Editable bool // not saved in database
	Created  time.Time
	Sublist  []Comment
}

type IndexData struct {
	Comments []Comment
}

var stopProcess chan bool

//var comments []Comment

var running bool = true
var port string
var tpl *template.Template
var database *sql.DB

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if running {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func initDB() {
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

	database = sqliteDatabase

	createCommentTable(sqliteDatabase)
	loadTestComments(sqliteDatabase)

	// DISPLAY INSERTED RECORDS
	//displayComments(sqliteDatabase, &comments)
}

func loadTestComments(db *sql.DB) {

	insertComment(db, "id-1", "test", "test message 1", "root", true, false)
	insertComment(db, "id-2", "test", "test message 2", "id-1", false, false)
	insertComment(db, "id-3", "test", "test message 3", "id-1", false, false)
	insertComment(db, "id-4", "test2", "test message 4", "id-3", false, false)
	insertComment(db, "id-5", "test2", "test message 5", "root", true, false)
	insertComment(db, "id-6", "test", "test message 6", "id-5", false, false)
	insertComment(db, "id-7", "test2", "test message 7", "id-5", false, false)
	insertComment(db, "id-8", "test", "test mail message 1", "test2", true, false)
	insertComment(db, "id-9", "test2", "test mail message 2", "test", true, false)

	//test child comment logic.
	getChildComments(db, "id-1", "test")
}

func getRootComments(db *sql.DB, username string) *[]Comment {
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
	rows, err := db.Query(`SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE root = 1 and parent = 'root'
						   ORDER BY created_at;`)

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

func getMailComments(db *sql.DB, parentID string, username string) *[]Comment {
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
	rows, err := db.Query(`WITH RECURSIVE descendants AS (
    					   SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE id = '` + parentID + `'
    
    					   UNION ALL
    
    					   SELECT m.id, m.user, m.message, m.picture, m.parent, m.root, m.sticky, m.created_at
    					   FROM comment m
    					   INNER JOIN descendants d ON m.parent = d.id
						   )
						   SELECT * FROM descendants
						   ORDER BY created_at;`)

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

func getChildComments(db *sql.DB, parentID string, username string) *[]Comment {
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
	rows, err := db.Query(`WITH RECURSIVE descendants AS (
    					   SELECT id, user, message, picture, parent, root, sticky, created_at
    					   FROM comment
    					   WHERE id = '` + parentID + `'
    
    					   UNION ALL
    
    					   SELECT m.id, m.user, m.message, m.picture, m.parent, m.root, m.sticky, m.created_at
    					   FROM comment m
    					   INNER JOIN descendants d ON m.parent = d.id
						   )
						   SELECT * FROM descendants
						   ORDER BY created_at;`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &user, &message, &picture, &parent, &root, &sticky, &created)
		editable = (username == user)
		log.Println("Comment ID:", id, " Message:", message, " Parent", parent, " Editable:", editable)
		if parent == parentID || root {
			comments = append(comments, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
		} else {
			log.Println("adding ", id, " to parent ", parent)
			AddCommentToSublist(&comments, parent, Comment{Id: id, User: user, Message: message, Picture: picture, Root: root, Sticky: sticky, Editable: editable, Created: created})
		}
	}

	return &comments

}

func createCommentTable(db *sql.DB) {
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
	"user" TEXT,
	"message" TEXT,
	"picture" TEXT,
	"parent" TEXT,
	"root" BIT,
	"sticky" BIT,
	"created_at" DATETIME
	);`

	log.Println("Create comment table...")
	statement, err := db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("comment table created")
}

func insertComment(db *sql.DB, id, user, message, parent string, root bool, sticky bool) {
	currentTime := time.Now()

	message = strings.Replace(message, "\n", "<br>", -1)

	log.Println("Inserting comment record ...")
	insertCommentSQL := `INSERT INTO comment(id, user, message,picture, parent,root,sticky,created_at) VALUES (?, ?, ?, ?, ?, ?, ?,?)`
	statement, err := db.Prepare(insertCommentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, user, message, "", parent, root, sticky, currentTime)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func editComment(db *sql.DB, id, message, parent string, root bool, sticky bool) {
	//currentTime := time.Now()

	message = strings.Replace(message, "\n", "<br>", -1)

	log.Println("Editing comment record ...")
	//insertCommentSQL := `INSERT INTO comment(id, user, message, parent,root,sticky,created_at) VALUES (?, ?, ?, ?, ?, ?,?)`

	editCommentSQL := `UPDATE comment SET message = ? WHERE id =?`
	statement, err := db.Prepare(editCommentSQL) // Prepare statement.
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

func editCommentPic(db *sql.DB, id, picture string) {
	//currentTime := time.Now()

	log.Println("Editing comment record to change picture ...")
	//insertCommentSQL := `INSERT INTO comment(id, user, message, parent,root,sticky,created_at) VALUES (?, ?, ?, ?, ?, ?,?)`

	editCommentSQL := `UPDATE comment SET picture = ? WHERE id =?`
	statement, err := db.Prepare(editCommentSQL) // Prepare statement.
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
func AddCommentToSublist(comments *[]Comment, id string, newComment Comment) bool {
	for i, comment := range *comments {
		if comment.Id == id {
			// If found, add newItem to the sublist
			(*comments)[i].Sublist = append((*comments)[i].Sublist, newComment)
			return true
		}
		// Recursive call to search in sublists
		if AddCommentToSublist(&comment.Sublist, id, newComment) {
			return true
		}
	}
	return false
}

// Helper function to print items and their sublists
func PrintComments(comments *[]Comment, indent string) {
	for _, comment := range *comments {
		fmt.Println(indent+"Comment:", comment.Message)
		if len(comment.Sublist) > 0 {
			PrintComments(&comment.Sublist, indent+"  ") // Increase indentation for sublists
		}
	}

}

func init() {

	port = ":8080"
	tpl = template.Must(template.ParseGlob("templates/*"))
	initDB()
	uuidWithHyphen := uuid.New()
	fmt.Println(uuidWithHyphen)

	//	handlers.Directory = "./downloads/"
}

/*
func handler(w http.ResponseWriter, r *http.Request) {

		log.Println("printing comments")
		PrintComments(&comments, "")
		tmpl := template.Must(template.ParseFiles("templates/collapse.html"))
		tmpl.Execute(w, comments)
	}
*/
func addHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	insertComment(database, id, username, message, parent, bRoot, bSticky)
	/*
		AddCommentToSublist(&comments, parent, comment)
		log.Println("added to sublist")
	*/

	log.Println("printing comments from:", r.FormValue("root"))
	currentComments := getChildComments(database, r.FormValue("root"), username)

	PrintComments(currentComments, "")

	tpl = template.Must(template.ParseFiles("templates/collapse.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", currentComments)
}

func editHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	id := r.FormValue("id")

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	editComment(database, id, message, parent, bRoot, bSticky)
	/*
		AddCommentToSublist(&comments, parent, comment)
		log.Println("added to sublist")
	*/

	log.Println("printing comments from:", r.FormValue("root"))
	currentComments := getChildComments(database, r.FormValue("root"), username)

	PrintComments(currentComments, "")

	tpl = template.Must(template.ParseFiles("templates/collapse.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", currentComments)
}

func indexEditHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	id := r.FormValue("id")

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	editComment(database, id, message, parent, bRoot, bSticky)

	log.Println("In index, user:", username)
	currentComments := getRootComments(database, username)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", currentComments)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {

	parent := r.PathValue("id")
	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	log.Println("In getCommentHandler looking for Parent:", parent)

	//currentComments := getChildComments(database, parent)

	//test child comment logic.
	currentComments := getChildComments(database, parent, username)

	tmpl := template.Must(template.ParseFiles("templates/collapse.html"))
	tmpl.Execute(w, currentComments)

}

func mailAddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	insertComment(database, id, username, message, parent, bRoot, bSticky)
	/*
		AddCommentToSublist(&comments, parent, comment)
		log.Println("added to sublist")
	*/

	log.Println("printing comments from:", r.FormValue("root"))
	currentComments := getMailComments(database, r.FormValue("root"), username)
	log.Println("message:", r.FormValue("comment"))

	PrintComments(currentComments, "")

	tpl = template.Must(template.ParseFiles("templates/mail.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", currentComments)
}

func mailHandler(w http.ResponseWriter, r *http.Request) {

	parent := r.PathValue("id")
	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	log.Println("In mailHandler looking for Parent:", parent)

	//currentComments := getChildComments(database, parent)

	//test child comment logic.
	currentComments := getMailComments(database, parent, username)

	tmpl := template.Must(template.ParseFiles("templates/mail.html"))
	tmpl.Execute(w, currentComments)

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	username := ""
	id := ""
	root := ""
	filename := ""
	source := ""

	// Ensure the request is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Grab the request's MultipartReader
	reader, err := r.MultipartReader()
	if err != nil {
		log.Println("multipart error:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Process the parts
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("error reading part:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if part.FileName() != "" {
			// Create the destination file
			dst, err := os.Create("./downloads/" + part.FileName())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filename = part.FileName()
			defer dst.Close()

			// Copy the part to dst
			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Process hidden fields (assuming they have a name)
			if part.FormName() == "X-User" {
				data, err := io.ReadAll(part)
				if err != nil {
					log.Println("error reading hidden field:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("X-User field value:", string(data))
				username = string(data)
			}
			if part.FormName() == "id" {
				data, err := io.ReadAll(part)
				if err != nil {
					log.Println("error reading hidden field:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("id field value:", string(data))
				id = string(data)
			}
			if part.FormName() == "root" {
				data, err := io.ReadAll(part)
				if err != nil {
					log.Println("error reading hidden field:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("root field value:", string(data))
				root = string(data)
			}
			if part.FormName() == "source" {
				data, err := io.ReadAll(part)
				if err != nil {
					log.Println("error reading hidden field:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("source field value:", string(data))
				source = string(data)
			}
		}
	}

	log.Println("in upload Handler")

	if username == "" {
		username = "test"
	}

	dirPath := "./downloads/" + username

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create the directory with 0755 permissions
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			//panic(err)
			fmt.Println("Error creating file directory:", err)
		}
		println("Directory created successfully.")
	} else if err != nil {
		panic(err)
	} else {
		println("Directory already exists.")
	}

	err = os.Rename("./downloads/"+filename, "./downloads/"+username+"/"+filename)
	if err != nil {
		fmt.Println("Error moving file:", err)
	} else {
		fmt.Println("File moved successfully.")
	}

	log.Println("uploading file from:", username, " adding to comment id:", id, " root Id:", root)
	editCommentPic(database, id, username+"/"+filename)

	switch source {
	case "comment":
		currentComments := getChildComments(database, root, username)
		PrintComments(currentComments, "")
		tpl = template.Must(template.ParseFiles("templates/collapse.html"))
		tpl.ExecuteTemplate(w, "comment-list-element", currentComments)
	case "mail":
		currentComments := getMailComments(database, root, username)
		tmpl := template.Must(template.ParseFiles("templates/mail.html"))
		tmpl.ExecuteTemplate(w, "comment-list-element", currentComments)
	case "index":
		currentComments := getRootComments(database, username)
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.ExecuteTemplate(w, "comment-list-element", currentComments)
	}

}

func indexAddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := "root"
	bRoot = true

	if username == "" {
		username = "test"
	}

	insertComment(database, id, username, message, parent, bRoot, bSticky)

	log.Println("In index, user:", username)
	currentComments := getRootComments(database, username)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", currentComments)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test"
	}

	log.Println("In index, user:", username)
	currentComments := getRootComments(database, username)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, currentComments)

}

func main() {

	stopProcess = make(chan bool)

	server := &http.Server{
		Addr: port,
	}

	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("./downloads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/healthcheck", healthCheck)

	http.HandleFunc("/", indexHandler)

	http.HandleFunc("/indexAddComment", indexAddHandler)
	http.HandleFunc("/collapseadd", addHandler)
	http.HandleFunc("/mailadd", mailAddHandler)

	http.HandleFunc("/indexEditComment", indexEditHandler)
	http.HandleFunc("/collapseedit", editHandler)

	http.HandleFunc("/comment/{id}/", commentHandler)
	http.HandleFunc("/mail/{id}/", mailHandler)

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Starting server on port: " + port)

	defer database.Close()

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	stopProcess <- true

	log.Println("Shutting Down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

}
