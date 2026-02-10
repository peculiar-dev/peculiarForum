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
	"strings"
	"time"

	"github.com/google/uuid"
)

type IndexHandler struct {
	comments data.Commentdb
	users    data.Userdb
	pageSize int
}

type IndexData struct {
	Comments     *[]data.Comment
	User         *data.User
	Page         int
	CurrentImage string
}

func NewIndexHandler(commentdb data.Commentdb, userdb data.Userdb, pageSize int) *IndexHandler {
	return &IndexHandler{comments: commentdb, users: userdb, pageSize: pageSize}
}

func (index *IndexHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test"
	}

	currentUser := index.users.GetUser(username)
	if currentUser.Username == "" {
		log.Println("User not found, calling userInit()")
		index.users.InsertUser(&data.User{Username: username})
		currentUser = index.users.GetUser(username)
	}

	log.Println("In index, user:", username)

	//currentComments := index.comments.GetRootComments(username)
	currentComments := index.comments.GetCommentsFromTo(username, 0, index.pageSize-1)

	data := IndexData{Comments: currentComments, User: currentUser, Page: 1, CurrentImage: ""}

	//tmpl := template.Must(template.ParseFiles("templates/header.html", "templates/index.html"))
	//tmpl.ExecuteTemplate(w, "index", currentComments)
	tmpl, err := template.ParseFiles("templates/header.html", "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//tmpl.Execute(w, currentComments)

}

func (index *IndexHandler) IndexPageHandler(w http.ResponseWriter, r *http.Request) {

	//page := r.PathValue("page")
	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	page, err := strconv.Atoi(r.PathValue("page"))
	if err != nil {
		log.Println("No page data, serving page 0.")
	}

	start := (page * index.pageSize)
	end := start + index.pageSize - 1

	currentComments := index.comments.GetCommentsFromTo(username, start, end)
	currentUser := index.users.GetUser(username)

	data := IndexData{Comments: currentComments, Page: page + 1, User: currentUser, CurrentImage: ""}

	log.Println("In page index, user:", username, " page:", page)
	//currentComments := index.comments.GetRootComments(username)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-item", data)
}

func (index *IndexHandler) AddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool
	//var commentLink string

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	linkAddr := r.FormValue("linkAddr")
	picture := r.FormValue("picture")

	parent := "root"
	bRoot = true
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		log.Println("No page in page add, serving 0.")
	}

	if username == "" {
		username = "test"
	}

	pics := strings.Split(picture, "/")
	pic := pics[len(pics)-2] + "/" + pics[len(pics)-1] // strip off user/pic

	log.Println("picture in Add:", pic)

	comment := data.Comment{Id: id, User: username, Message: message, Parent: parent, Root: bRoot, Sticky: bSticky, Picture: pic, Link: linkAddr}
	index.comments.InsertComment(comment)

	//index.comments.InsertComment(id, "", username, message, linkAddr, parent, bRoot, bSticky)

	log.Println("In add index, user:", username, " link:", linkAddr)
	//currentComments := index.comments.GetRootComments(username)
	start := 0
	end := (page * index.pageSize) - 1

	currentComments := index.comments.GetCommentsFromTo(username, start, end)

	//currentComments := index.comments.GetCommentsFromTo(username, start, end)
	currentUser := index.users.GetUser(username)

	data := IndexData{Comments: currentComments, User: currentUser, Page: page, CurrentImage: ""}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", data)
}

func (index *IndexHandler) EditHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	id := r.FormValue("id")

	linkAddr := r.FormValue("linkAddr")

	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		log.Println("No page in page edit, serving 0.")
	}

	log.Println("sticky:", r.FormValue("sticky"))
	log.Println("id:", id)
	log.Println("page:", page)

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	currentUser := index.users.GetUser(username)
	currentComment := index.comments.GetComment(id)
	if currentUser.Level == 100 {
		log.Println("User is Admin")
		if r.FormValue("sticky") == "true" {
			bSticky = true
		}
	} else {
		bSticky = currentComment.Sticky
	}

	updateTime := currentComment.Created

	// if current comment was not sticky, but is now sticky, add 30 days to now
	if !currentComment.Sticky && bSticky {
		updateTime = updateTime.Add(30 * 24 * time.Hour) // now + 30 days
	} else {
		updateTime = time.Now()
	}
	// if current comment was sticky, but is now not sticky, just set it to today.

	index.comments.EditComment(id, message, linkAddr, parent, bRoot, bSticky, updateTime)

	log.Println("In edit index, user:", username)
	//currentComments := index.comments.GetRootComments(username)
	start := 0
	end := (page * index.pageSize) - 1

	currentComments := index.comments.GetCommentsFromTo(username, start, end)

	data := IndexData{Comments: currentComments, User: currentUser, Page: page, CurrentImage: ""}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", data)
}

func (index *IndexHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	id := ""
	root := ""
	filename := ""
	page := 0

	username := extractBasicAuthUsername(r)

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
			if part.FormName() == "page" {
				data, err := io.ReadAll(part)
				if err != nil {
					log.Println("error reading hidden field:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				log.Println("page field value:", string(data))
				page, err = strconv.Atoi(string(data))
				if err != nil {
					log.Println("Error, invalid page in IndexPageHandler.")
				}

			}
		}
	}

	log.Println("in index upload Handler")

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
	index.comments.EditCommentPic(id, username+"/"+filename)

	currentUser := index.users.GetUser(username)
	//currentComments := index.comments.GetRootComments(username)
	start := 0
	end := (page * index.pageSize) - 1

	currentComments := index.comments.GetCommentsFromTo(username, start, end)

	data := IndexData{Comments: currentComments, User: currentUser, Page: page, CurrentImage: ""}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", data)

}

func (index *IndexHandler) NewPostUploadHandler(w http.ResponseWriter, r *http.Request) {

	id := ""
	//root := ""
	filename := ""
	//page := 0

	username := extractBasicAuthUsername(r)

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
			/*
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
				if part.FormName() == "page" {
					data, err := io.ReadAll(part)
					if err != nil {
						log.Println("error reading hidden field:", err.Error())
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					log.Println("page field value:", string(data))
					page, err = strconv.Atoi(string(data))
					if err != nil {
						log.Println("Error, invalid page in IndexPageHandler.")
					}
				}
			*/
		}
	}

	log.Println("in index upload Handler")

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

	log.Println("uploading file from:", username, " and setting as current file")
	index.comments.EditCommentPic(id, username+"/"+filename)

	//currentUser := index.users.GetUser(username)
	//currentComments := index.comments.GetRootComments(username)
	//start := 0
	//end := (page * index.pageSize) - 1

	//rrentComments := index.comments.GetCommentsFromTo(username, start, end)

	data := IndexData{CurrentImage: "./downloads/" + username + "/" + filename}

	//tmpl := template.Must(template.ParseFiles("templates/index.html"))
	//tmpl.ExecuteTemplate(w, "comment-list-element", data)

	//make sure the page shows new data, not cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	tmpl, err := template.ParseFiles("templates/header.html", "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "sub-pic-element", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
