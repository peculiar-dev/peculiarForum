package main

import (
	"context"
	"peculiarity/internal/data"
	"peculiarity/internal/handlers"

	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	Id       string    // the UUID of this comment
	User     string    // The Username of the creator
	Message  string    // The comment message
	Picture  string    // The picture uploaded with this comment
	Root     bool      // Is this a root comment?
	Sticky   bool      // Is this a 'sticky' or 'pinned' comment?
	Editable bool      // not saved in database
	Created  time.Time // Creation timestamp
	Sublist  []Comment // Children of this comment
}

type IndexData struct {
	Comments []Comment
}

var commentsdb data.Commentdb
var userdb data.Userdb
var notificationdb data.Notificationdb

var stopProcess chan bool

//var comments []Comment

var running bool = true
var port string

//var tpl *template.Template

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if running {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func init() {

	port = ":8080"
	//tpl = template.Must(template.ParseGlob("templates/*"))

	commentsdb = data.NewSqliteCommentDB()
	commentsdb.InitDB()

	userdb = data.NewSqliteUserDB()
	userdb.Setdb(commentsdb.Getdb()) // set the userdb to the same sqlite db instance
	userdb.CreateUserTable()
	userdb.LoadTestUsers()
	log.Default().Println(userdb.GetUsers())

	notificationdb = data.NewSqliteNotificationDB()
	notificationdb.Setdb(commentsdb.Getdb()) // set the notification db to the same sqline db instance
	notificationdb.CreateNotificationTable()
	notificationdb.LoadTestNotifications()
	log.Default().Println(notificationdb.GetNotifications("test"))

	uuidWithHyphen := uuid.New()
	fmt.Println(uuidWithHyphen)

	//	handlers.Directory = "./downloads/"
}

func main() {

	stopProcess = make(chan bool)

	server := &http.Server{
		Addr: port,
	}

	indexhandler := handlers.NewIndexHandler(commentsdb, userdb)
	mailhandler := handlers.NewMailHandler(commentsdb, userdb)
	commentHandler := handlers.NewCommentHandler(commentsdb, notificationdb)
	notificationHandler := handlers.NewNotificationHandler(notificationdb)
	userHandler := handlers.NewUserHandler(userdb)

	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("./downloads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/healthcheck", healthCheck)

	//comment index

	http.HandleFunc("/", indexhandler.IndexHandler)
	http.HandleFunc("/indexAddComment", indexhandler.AddHandler)
	http.HandleFunc("/indexEditComment", indexhandler.EditHandler)
	http.HandleFunc("/indexUpload", indexhandler.UploadHandler)

	//mail index

	http.HandleFunc("/mailIndex", mailhandler.IndexHandler)
	http.HandleFunc("/indexAddHandler", mailhandler.IndexAddHandler)
	http.HandleFunc("/mailadd", mailhandler.AddHandler)
	http.HandleFunc("/mail/{id}/", mailhandler.IDHandler)
	http.HandleFunc("/mailUpload", mailhandler.UploadHandler)

	//collapsable comment view

	http.HandleFunc("/collapseadd", commentHandler.AddHandler)
	http.HandleFunc("/collapseedit", commentHandler.EditHandler)
	http.HandleFunc("/comment/{id}/", commentHandler.IDHandler)
	http.HandleFunc("/commentUpload", commentHandler.UploadHandler)

	// notifications

	http.HandleFunc("/notifications", notificationHandler.IndexHandler)

	// user

	http.HandleFunc("/user", userHandler.IndexHandler)
	http.HandleFunc("/userUpdate", userHandler.UpdateHandler)

	log.Println("Starting server on port: " + port)

	defer commentsdb.Getdb().Close()
	defer userdb.Getdb().Close()

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
