package main

import (
	"context"
	"os"
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
var port string = ":8080"
var debug bool = false
var initialize bool = false

//var tpl *template.Template

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if running {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func init() {

	// handle command-line arguments
	if len(os.Args) > 1 {
		for i, arg := range os.Args[1:] {
			if arg == "-debug" {
				debug = true
				log.Println("setting debug mode true")
			}
			if arg == "-init" {
				initialize = true
				log.Println("setting init mode to true")
			}
			if arg == "-port" {
				port = ":" + os.Args[i+2]
				log.Println("setting port to:", port)
			}
			if arg == "-h" || arg == "-help" {
				fmt.Println(" -h or -help         - This help message.")
				fmt.Println(" -debug              - Sets debug mode to true, default false.")
				fmt.Println(" -init               - Sets initalize mode to true, default false.")
				fmt.Println(" -port <port number> - Sets the port number to <port number>, default: 8080")
				os.Exit(0)
			}

		}
	}

	//tpl = template.Must(template.ParseGlob("templates/*"))

	commentsdb = data.NewSqliteCommentDB()
	commentsdb.InitDB(initialize, debug)

	userdb = data.NewSqliteUserDB()
	userdb.Setdb(commentsdb.Getdb()) // set the userdb to the same sqlite db instance
	if initialize {
		userdb.CreateUserTable()
	}
	if debug {
		userdb.LoadTestUsers()
		log.Default().Println(userdb.GetUsers())
	}

	notificationdb = data.NewSqliteNotificationDB()
	notificationdb.Setdb(commentsdb.Getdb()) // set the notification db to the same sqline db instance
	if initialize {
		notificationdb.CreateNotificationTable()
	}
	if debug {
		notificationdb.LoadTestNotifications()
		log.Default().Println(notificationdb.GetNotifications("test"))
	}

	uuidWithHyphen := uuid.New()
	fmt.Println(uuidWithHyphen)

	//	handlers.Directory = "./downloads/"
}

func main() {

	stopProcess = make(chan bool)

	server := &http.Server{
		Addr: port,
	}

	indexhandler := handlers.NewIndexHandler(commentsdb, userdb, 10)
	mailhandler := handlers.NewMailHandler(commentsdb, userdb, notificationdb)
	commentHandler := handlers.NewCommentHandler(commentsdb, notificationdb, userdb)
	notificationHandler := handlers.NewNotificationHandler(notificationdb, userdb)
	userHandler := handlers.NewUserHandler(userdb)
	chatHandler := handlers.NewChatHandler(userdb)

	http.Handle("/user/", http.StripPrefix("/user/", http.FileServer(http.Dir("./downloads/user"))))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("./downloads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/healthcheck", healthCheck)

	//comment index

	http.HandleFunc("/", indexhandler.IndexHandler)
	http.HandleFunc("/{page}/", indexhandler.IndexPageHandler)
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

	http.HandleFunc("/settings", userHandler.IndexHandler)
	http.HandleFunc("/userUpdate", userHandler.UpdateHandler)
	http.HandleFunc("/userLevelUpdate", userHandler.UpdateLevelHandler)
	http.HandleFunc("/userIconUpload", userHandler.UploadPhotoHandler)
	http.HandleFunc("/userFileUpload", userHandler.UploadFileHandler)
	http.HandleFunc("/userFileDelete/{filename}/", userHandler.FileDeleteHandler)

	//chat
	http.HandleFunc("/chat", chatHandler.ChatIndexHandler)
	http.HandleFunc("/chatSocket", chatHandler.ChatSocket)

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
