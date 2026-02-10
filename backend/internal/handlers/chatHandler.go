package handlers

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"peculiarity/internal/data"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type ChatUser struct {
	Name       string
	Connection *websocket.Conn
}

type ChatHandler struct {
	users   data.Userdb
	clients []*ChatUser
}

type ChatData struct {
	User *data.User
}

var upgrader = websocket.Upgrader{} // use default options

func NewChatHandler(userdb data.Userdb) *ChatHandler {
	return &ChatHandler{users: userdb}
}

func (chat *ChatHandler) ChatIndexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test"
	}
	currentUser := chat.users.GetUser(username)

	data := ChatData{User: currentUser}

	tmpl, err := template.ParseFiles("templates/header.html", "templates/chat.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "chat.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// this function will just echo the incoming data to the appropriate plugin
func (chat *ChatHandler) ChatSocket(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool {
		log.Print(r)

		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		return equalASCIIFold(u.Host, r.Host)
	}
	c, err := upgrader.Upgrade(w, r, nil)

	c.SetPongHandler(func(appData string) error {
		log.Println("pong", appData)
		return nil
	})

	go func() {
		for {
			time.Sleep(10 * time.Second)
			log.Println("ping")
			err := c.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("Failed to send ping:", err)
				return
			}
		}
	}()

	var user ChatUser
	user.Connection = c
	chat.clients = append(chat.clients, &user)
	log.Print("Clients Connected:", len(chat.clients))
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for { // infinite read loop.
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		for idx, user := range chat.clients {
			log.Println("idx", idx, " User:", user)
			if user != nil {
				if user.Connection != nil {
					err := user.Connection.WriteMessage(websocket.TextMessage, []byte(message))
					//err = c.WriteMessage(mt, message)
					if err != nil {
						log.Println("err:", err)

						if user.Connection != nil {
							user.Connection.Close()
							user.Name = "_remove"

						}

					}
				}
			}
		}
		log.Println("clearning up")
		var temp []*ChatUser
		for idx, user := range chat.clients {
			if user.Name != "_remove" {
				temp = append(temp, user)
			} else {
				log.Println("debug:", "removing user:", idx)
			}
		}
		chat.clients = temp
	}

}

func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
