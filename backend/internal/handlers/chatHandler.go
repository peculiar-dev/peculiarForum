package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"peculiarity/internal/data"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type ChatUser struct {
	Name       string
	Connection *websocket.Conn
}

type ChatHandler struct {
	users         data.Userdb
	clients       []*ChatUser
	buffer        []bufferedMessage
	inactiveUsers map[string]time.Time
}

type ChatData struct {
	User *data.User
}

type bufferedMessage struct {
	Message   string
	Timestamp time.Time
}

var upgrader = websocket.Upgrader{} // use default options

func NewChatHandler(userdb data.Userdb) *ChatHandler {
	return &ChatHandler{users: userdb, buffer: make([]bufferedMessage, 0, 100), inactiveUsers: make(map[string]time.Time)}
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

func (chat *ChatHandler) ChatUnreadCountHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test"
	}
	//currentUser := chat.users.GetUser(username)
	unreadChatCount := 0

	for _, message := range chat.buffer {
		if message.Timestamp.After(chat.inactiveUsers[username]) {
			unreadChatCount++
		}
	}

	fmt.Fprintf(w, "<span class=\"chat-unread\">%d</span>", unreadChatCount)

}

func (chat *ChatHandler) ParseChatCommand(command string, user *ChatUser) error {
	// split the command into parts
	parts := strings.Fields(command)
	switch parts[0] {
	case "/list":
		chat.listUsers(user.Connection)
		return nil
	case "/help":
		chat.help(user.Connection)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}
}

func (chat *ChatHandler) listUsers(WebSocket *websocket.Conn) {
	var userList []string
	for _, user := range chat.clients {
		if user.Name != "_remove" {
			userList = append(userList, user.Name)
		}
	}
	userListMessage := "Users in chat: <br>" + strings.Join(userList, "<br>")
	err := WebSocket.WriteMessage(websocket.TextMessage, []byte(userListMessage))
	if err != nil {
		log.Println("Failed to send user list:", err)
	}
}

func (chat *ChatHandler) help(WebSocket *websocket.Conn) {
	helpMessage := `Available commands: <br>
	                /list - List users in chat<br>
					/help - Show this help message`

	err := WebSocket.WriteMessage(websocket.TextMessage, []byte(helpMessage))
	if err != nil {
		log.Println("Failed to send help message:", err)
	}
}

// this function will just echo the incoming data to the appropriate plugin
func (chat *ChatHandler) ChatSocket(w http.ResponseWriter, r *http.Request) {

	var user ChatUser
	user.Name = r.Header.Get("X-User")
	if user.Name == "" {
		user.Name = "test"
	}

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
		//log.Println("pong", appData)
		return nil
	})

	go func() {
		for {
			time.Sleep(10 * time.Second)
			//log.Println("ping")
			err := c.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("Failed to send ping:", err)
				chat.removeUser(&user)
				return
			}
		}
	}()

	user.Connection = c
	chat.clients = append(chat.clients, &user)
	log.Print("Clients Connected:", len(chat.clients))
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	log.Println("sending buffered messages to new client")
	for _, bufferedMessage := range chat.buffer {
		err := c.WriteMessage(websocket.TextMessage, []byte(bufferedMessage.Message))
		if err != nil {
			log.Println("Failed to send buffered message:", err)
			return
		}
	}
	err = c.WriteMessage(websocket.TextMessage, []byte("Welcome to chat, type /help for help."))
	if err != nil {
		log.Println("Failed to send welcome message:", err)
		return
	}
	for { // infinite read loop.
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		if len(message) > 0 && message[0] == '/' {
			err := chat.ParseChatCommand(string(message), &user)
			if err != nil {
				log.Println("Failed to parse command:", err)
				continue
			}
			continue
		} else {
			chat.addToBuffer(string(message))
			for idx, user := range chat.clients {
				log.Println("idx", idx, " User:", user.Name)
				if user != nil {
					if user.Connection != nil {
						err := user.Connection.WriteMessage(websocket.TextMessage, []byte(message))
						//err = c.WriteMessage(mt, message)
						if err != nil {
							log.Println("err:", err)
							chat.removeUser(user)
							if user.Connection != nil {
								user.Connection.Close()
							}
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

func (chat *ChatHandler) removeUser(user *ChatUser) {
	chat.inactiveUsers[user.Name] = time.Now()
	log.Println("marked user as inactive:", user.Name)
	for n, t := range chat.inactiveUsers {
		log.Println("inactive user:", n, t)
	}
	user.Name = "_remove"
}

func (chat *ChatHandler) addToBuffer(message string) {

	if len(chat.buffer) < 100 {
		chat.buffer = append(chat.buffer, bufferedMessage{Message: message, Timestamp: time.Now()})
	} else {
		copy(chat.buffer, chat.buffer[1:])
		chat.buffer[len(chat.buffer)-1] = bufferedMessage{Message: message, Timestamp: time.Now()}
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
