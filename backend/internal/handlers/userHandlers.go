package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"peculiarity/internal/data"
	"strconv"
)

type UserHandler struct {
	users data.Userdb
}

type UserIndexData struct {
	User   *data.User
	Themes *[]string
	Files  []os.FileInfo
	Users  *[]data.User
}

func NewUserHandler(userdb data.Userdb) *UserHandler {
	return &UserHandler{users: userdb}
}

func (userHandler *UserHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	var users *[]data.User

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test" // set a test user
	}
	log.Println("In user settings, user:", username)

	currentUser := userHandler.users.GetUser(username)

	log.Println("user level:", currentUser.Level)

	if currentUser.Level >= 100 {
		log.Println("Getting users.")
		users = userHandler.users.GetUsers()
	}

	//themes := []string{"dark", "light"}
	themes := getThemes()
	files := ListFile(username)
	indexData := UserIndexData{currentUser, themes, files, users}

	log.Println("User has name:" + currentUser.Username)
	log.Println("User has theme:" + currentUser.Theme)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

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

	var users *[]data.User

	username := r.Header.Get("X-User")
	theme := r.FormValue("theme")
	email := r.FormValue("email")

	if username == "" {
		username = "test"
	}

	currentUser := userHandler.users.GetUser(username)

	if currentUser.Level >= 100 {
		users = userHandler.users.GetUsers()
	}

	currentUser.Theme = theme
	currentUser.Email = email

	userHandler.users.UpdateUser(currentUser)

	log.Println("In update user:", username)

	themes := getThemes()
	files := ListFile(username)
	indexData := UserIndexData{currentUser, themes, files, users}

	log.Println("User has theme:" + currentUser.Theme)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

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

func (userHandler *UserHandler) UpdateLevelHandler(w http.ResponseWriter, r *http.Request) {

	var users *[]data.User

	username := r.Header.Get("X-User")
	updateName := r.FormValue("user")
	//level := r.FormValue("level")
	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		log.Println("Error, invalid level in UpdateHandler.")
	}

	if username == "" {
		username = "test"
	}

	log.Println("In update user Level user:", username, " is updating:", updateName)

	currentUser := userHandler.users.GetUser(username)

	if currentUser.Level >= 100 {
		users = userHandler.users.GetUsers()
		updateUser := userHandler.users.GetUser(updateName)
		updateUser.Level = level
		log.Println("updating:", updateName, " level to:", level)
		userHandler.users.UpdateUser(updateUser)
		updateUser = userHandler.users.GetUser(updateName)
		log.Println("user:", updateName, " level now set to:", updateUser.Level)
	}

	themes := getThemes()
	files := ListFile(username)
	indexData := UserIndexData{currentUser, themes, files, users}

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "level-element", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (userHandler *UserHandler) UploadPhotoHandler(w http.ResponseWriter, r *http.Request) {

	var users *[]data.User

	username := ""
	filename := "_user_icon.png"

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
			dst, err := os.Create("./downloads/" + filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//filename = part.FileName()
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

	//display the sub-template

	if username == "" {
		username = "test"
	}

	currentUser := userHandler.users.GetUser(username)

	themes := getThemes()
	files := ListFile(username)
	indexData := UserIndexData{currentUser, themes, files, users}

	log.Println("User has theme:" + currentUser.Theme)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "icon-element", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (userHandler *UserHandler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {

	var users *[]data.User

	username := ""
	filename := ""

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
			filename = part.FileName()
			dst, err := os.Create("./downloads/" + filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//filename = part.FileName()
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

		}
	}

	log.Println("in File upload Handler")

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

	err = os.Rename("./downloads/"+filename, "./downloads/user/"+username+"/"+filename)
	if err != nil {
		fmt.Println("Error moving file:", err)
	} else {
		fmt.Println("File moved successfully.")
	}

	//display the sub-template

	if username == "" {
		username = "test"
	}

	currentUser := userHandler.users.GetUser(username)

	themes := getThemes()
	files := ListFile(username)
	indexData := UserIndexData{currentUser, themes, files, users}

	log.Println("User has theme:" + currentUser.Theme)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "file-element", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (userHandler *UserHandler) FileDeleteHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	filename := r.PathValue("filename")
	currentUser := userHandler.users.GetUser(username)

	userHandler.users.UpdateUser(currentUser)

	log.Println("In delete user file:", username, " file:", filename)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	path := "./downloads/user/" + username + "/" + filename

	err := os.Remove(path)
	if err != nil {
		fmt.Println("Error deleting file:", err)
	} else {
		fmt.Println("File deleted successfully")
	}

	var indexData UserIndexData

	files := ListFile(username)
	indexData.Files = files
	indexData.User = currentUser

	tmpl, err := template.ParseFiles("templates/header.html", "templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "file-element", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ListFile(username string) []os.FileInfo {
	log.Println("Listing  files to template")
	f, err := os.Open("./downloads/user/" + username + "/")
	if err != nil {
		log.Print(err)
	}
	// read the whole directory
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Print(err)
	}
	onlyfiles := make([]os.FileInfo, 0)
	// ignore folders
	for _, file := range files {
		if !file.IsDir() {
			log.Println("File:", file.Name())
			onlyfiles = append(onlyfiles, file)
		}
	}
	return onlyfiles
}
