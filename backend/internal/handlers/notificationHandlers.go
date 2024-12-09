package handlers

import (
	"html/template"
	"log"
	"net/http"
	"peculiarity/internal/data"
)

type NotificationHandler struct {
	notifications data.Notificationdb
}

func NewNotificationHandler(notificationdb data.Notificationdb) *NotificationHandler {
	return &NotificationHandler{notifications: notificationdb}
}

func (notification *NotificationHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test2" // this is so you can reply on the comments page, and see notifications update.
		// note when testing without a login framework that you are testing with 'test'
		// as a user for other things, but 'test2' for notifications.
	}

	log.Println("In notifications, user:", username)
	currentNotifications := notification.notifications.GetNotifications(username)

	//tmpl := template.Must(template.ParseFiles("templates/header.html", "templates/index.html"))
	//tmpl.ExecuteTemplate(w, "index", currentComments)
	tmpl, err := template.ParseFiles("templates/header.html", "templates/notification.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "notification.html", currentNotifications)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//tmpl.Execute(w, currentComments)

}
