package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"peculiarity/internal/data"
	"peculiarity/internal/email"
	"peculiarity/internal/handlers"
	"strconv"
	"strings"
	"syscall"

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
var emailConnector email.EmailConnector
var updateHour int // the hour of the day when daily updates will run.
var notificationBody string
var chatBody string

var stopProcess chan os.Signal

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

func chatInvite(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")
	recipient := r.FormValue("user")

	inviteUser := userdb.GetUser(recipient)
	err := emailConnector.SendNotification(inviteUser.Email, username+" invited you to chat: "+chatBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func init() {

	updateHour = 19 // default 7pm

	loadConfig() // load config from file

	// handle commandline arguments (overrides commandline )
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
			if arg == "-uHour" { // update hour
				updateHour, err := strconv.Atoi(os.Args[i+2])
				if err != nil {
					log.Println("hour set by -uHour must be a number between 0 and 24")
					os.Exit(0)
				}
				if updateHour > 24 {
					log.Println("hour set by -uHour in command line must not be greater than 24")
					os.Exit(0)
				}
				if updateHour < 0 {
					log.Println("hour set by -uHour in command line must not be less than 0")
					os.Exit(0)
				}
				log.Println("setting update Hour to:", updateHour)
			}
			if arg == "-h" || arg == "-help" {
				fmt.Println(" -h or -help         - This help message.")
				fmt.Println(" -debug              - Sets debug mode to true, default false.")
				fmt.Println(" -init               - Sets initalize mode to true, default false.")
				fmt.Println(" -port <port number> - Sets the port number to <port number>, default: 8080")
				fmt.Println(" -uHour <hour from 0-24")
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
		userdb.LoadTestUsers()
		log.Default().Println(userdb.GetUsers())
	}

	notificationdb = data.NewSqliteNotificationDB()
	notificationdb.Setdb(commentsdb.Getdb()) // set the notification db to the same sqline db instance
	if initialize {
		notificationdb.CreateNotificationTable()
		notificationdb.LoadTestNotifications()
		log.Default().Println(notificationdb.GetNotifications("test"))
	}

	uuidWithHyphen := uuid.New()
	log.Println(uuidWithHyphen)

	//	handlers.Directory = "./downloads/"
}

func loadConfig() {
	filePath := "pforum.conf"
	var emailType, password, emailAddress, smtpHost string

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("error opening config file:", err)
		os.Exit(0)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Println("error reading config file:", err)
		os.Exit(0)
	}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			setting := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch setting {
			case "debug":
				debug, err = strconv.ParseBool(value)
			case "init":
				initialize, err = strconv.ParseBool(value)
			case "port":
				port = value
			case "uHour":
				updateHour, err := strconv.Atoi(value)
				if err != nil {
					log.Println("hour set by -uHour must be a number between 0 and 24")
					os.Exit(0)
				}
				if updateHour > 24 {
					log.Println("hour set by -uHour in command line must not be greater than 24")
					os.Exit(0)
				}
				if updateHour < 0 {
					log.Println("hour set by -uHour in command line must not be less than 0")
					os.Exit(0)
				}
			case "emailType":
				emailType = value
			case "password":
				password = value
			case "emailAddress":
				emailAddress = value
			case "smtpHost":
				smtpHost = value
			case "notificationBody":
				notificationBody = value
			case "chatBody":
				chatBody = value
			default:
				log.Printf("invalid setting in pforum.conf. Setting: %s, value: %s. setting not found.", setting, value)
				os.Exit(0)
			}
			if err != nil {
				log.Printf("invalid setting in pforum.conf. Setting: %s, value: %s caused error: %s /n", setting, value, err.Error())
				os.Exit(0)
			}
		}
	}
	if emailType == "SMTP" {
		emailConnector = email.NewEmailSMTP(smtpHost, emailAddress, password)
	}

}

func sendEmail() {
	users := userdb.GetUsers()

	for _, user := range *users {
		if notificationdb.HasNotifications(user.Username, user.LastLogin) && user.Email != "" {
			log.Println("Sending email to user: " + user.Username + " email: " + user.Email)
			//emailConnector.SendNotification(user.Email, "https://comopeculiarity.org/notifications")
			emailConnector.SendNotification(user.Email, notificationBody)
		}
	}
}

func dailyUpdates() {

	/*
		password = "zcqnhaqfpschjmmq" // generated 16 digits app password
		from     = "como.peculiarity@gmail.com"
		smtpHost = "smtp.gmail.com"
	*/

	//emailConnector = email.NewEmailSMTP("smtp.gmail.com", "como.peculiarity@gmail.com", "zcqnhaqfpschjmmq")

	for {
		now := time.Now()
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, updateHour, 0, 0, 0, time.UTC)
		fmt.Println("Daily updates will run tomorrow: ", tomorrow)
		timeTo := tomorrow.Sub(now)

		/*email each user who has notifications newer than their last login*/
		//fmt.Println("has updates:", notificationdb.HasNotifications("test", time.Now().Add(3*time.Hour)))

		sendEmail()

		time.Sleep(timeTo)
	}

}

func main() {

	stopProcess = make(chan os.Signal, 1)
	signal.Notify(stopProcess, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr: port,
	}

	indexhandler := handlers.NewIndexHandler(commentsdb, userdb, 10)
	mailhandler := handlers.NewMailHandler(commentsdb, userdb, notificationdb)
	commentHandler := handlers.NewCommentHandler(commentsdb, notificationdb, userdb)
	notificationHandler := handlers.NewNotificationHandler(notificationdb, userdb)
	userHandler := handlers.NewUserHandler(userdb)
	chatHandler := handlers.NewChatHandler(userdb)

	//http.Handle("/user/", http.StripPrefix("/user/", http.FileServer(http.Dir("./downloads/user"))))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("./downloads"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/healthcheck", healthCheck)
	http.HandleFunc("/chatInvite", chatInvite)

	//comment index

	http.HandleFunc("/", indexhandler.IndexHandler)
	http.HandleFunc("/{page}/", indexhandler.IndexPageHandler)
	http.HandleFunc("/indexAddComment", indexhandler.AddHandler)
	http.HandleFunc("/indexEditComment", indexhandler.EditHandler)
	http.HandleFunc("/indexUpload", indexhandler.UploadHandler)
	http.HandleFunc("/indexSubPicUpload", indexhandler.NewPostUploadHandler)

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

	go dailyUpdates()

	go func() {
		<-stopProcess
		log.Println("Shutting Down")

		shutdownCTX, shutdownRel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownRel()

		if err := server.Shutdown(shutdownCTX); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
		log.Println("HTTP Shutdown")

		log.Println("Shutting Down DB Connections")
		commentsdb.Getdb().Close()
		userdb.Getdb().Close()
		notificationdb.Getdb().Close()
		log.Println("DB Connections Shutdown")

		os.Exit(0)
	}()

	log.Println("Starting http Server.")
	if err := http.ListenAndServe(port, nil); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen error: %s\n", err)
	}

}
