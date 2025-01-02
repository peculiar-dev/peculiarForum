package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"peculiarity/internal/data"
)

type UserHandler struct {
	users data.Userdb
}

type UserIndexData struct {
	User   *data.User
	Themes *[]string
}

func NewUserHandler(userdb data.Userdb) *UserHandler {
	return &UserHandler{users: userdb}
}

func (userHandler *UserHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test" // set a test user
	}
	log.Println("In user settings, user:", username)

	currentUser := userHandler.users.GetUser(username)
	//themes := []string{"dark", "light"}
	themes := getThemes()
	indexData := UserIndexData{currentUser, themes}

	log.Println("User has name:" + currentUser.Username)
	log.Println("User has theme:" + currentUser.Theme)

	//tmpl := template.Must(template.ParseFiles("templates/header.html", "templates/index.html"))
	//tmpl.ExecuteTemplate(w, "index", currentComments)
	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "user.html", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func getThemes() *[]string {

	f, err := os.Open("./static/themes/")
	if err != nil {
		log.Print(err)
	}
	// read the whole directory
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Print(err)
	}
	onlydirs := make([]string, 0)
	// ignore files
	for _, file := range files {
		if file.IsDir() {
			onlydirs = append(onlydirs, file.Name())
		}
	}
	return &onlydirs

}

func (userHandler *UserHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	theme := r.FormValue("theme")

	if username == "" {
		username = "test"
	}

	currentUser := userHandler.users.GetUser(username)

	currentUser.Theme = theme

	userHandler.users.UpdateUser(currentUser)

	log.Println("In update user:", username)

	themes := getThemes()
	indexData := UserIndexData{currentUser, themes}

	log.Println("User has theme:" + currentUser.Theme)

	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "user.html", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
