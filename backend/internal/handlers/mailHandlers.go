package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"peculiarity/internal/data"

	"github.com/google/uuid"
)

type MailHandler struct {
	comments data.Commentdb
	users    data.Userdb
}

type MailIndexData struct {
	Comments *[]data.Comment
	Users    *[]data.User
	User     *data.User
}

func NewMailHandler(commentdb data.Commentdb, userdb data.Userdb) *MailHandler {
	return &MailHandler{comments: commentdb, users: userdb}
}

func (mail *MailHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	username := r.Header.Get("X-User")
	if username == "" {
		username = "test"
	}

	log.Println("In Mail index, user:", username)
	//currentComments := mail.comments.GetRootMail(username)
	var indexData MailIndexData
	indexData.Comments = mail.comments.GetRootMail(username)
	indexData.Users = mail.users.GetUsers()
	indexData.User = mail.users.GetUser(username)

	tmpl, err := template.ParseFiles("templates/header.html", "templates/mailIndex.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//err = tmpl.ExecuteTemplate(w, "mailIndex.html", currentComments)
	err = tmpl.ExecuteTemplate(w, "mailIndex.html", indexData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (mail *MailHandler) AddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	linkAddr := r.FormValue("linkAddr")

	if username == "" {
		username = "test"
	}

	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", parent)
	//fmt.Printf("comment:%v\n", comment)

	mail.comments.InsertComment(id, r.FormValue("root"), username, message, linkAddr, parent, bRoot, bSticky)
	//mail.comments.InsertComment(id, username, message, username+"-"+recipient, bRoot, bSticky)

	log.Println("printing mail comments from:", r.FormValue("root"))

	//currentComments := mail.comments.GetMailComments(r.FormValue("root"), username)
	var indexData MailIndexData
	indexData.Comments = mail.comments.GetMailComments(r.FormValue("root"), username)
	//indexData.Users = mail.users.GetUsers()
	//indexData.User = mail.users.GetUser(username)

	log.Println("message:", r.FormValue("comment"))

	tpl := template.Must(template.ParseFiles("templates/mail.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", indexData)
}

func (mail *MailHandler) IndexAddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	//parent := r.FormValue("parent")
	recipient := r.FormValue("user")
	linkAddr := r.FormValue("linkAddr")

	bRoot = true

	if username == "" {
		username = "test"
	}
	//root := r.FormValue("root") // make boolean?
	//sticky := r.FormValue("sticky") // make boolean?

	//parent = parent[10:] // strip javascript identifier
	//comment := Comment{Id: id, User: username, Message: message, Root: bRoot, Sticky: bSticky, Sublist: nil}

	fmt.Printf("parent: %s\n", username+"-"+recipient)
	//fmt.Printf("comment:%v\n", comment)

	//mail.comments.InsertComment(id, username, message, parent, bRoot, bSticky)
	mail.comments.InsertComment(id, "", username, message, linkAddr, username+"-"+recipient, bRoot, bSticky)

	/*
		log.Println("printing mail comments from:", r.FormValue("root"))
		currentComments := mail.comments.GetMailComments(r.FormValue("root"), username)
		log.Println("message:", r.FormValue("comment"))

		tpl := template.Must(template.ParseFiles("templates/mail.html"))

		tpl.ExecuteTemplate(w, "comment-list-element", currentComments)
	*/

	log.Println("In Mail index add, user:", username)
	//currentComments := mail.comments.GetRootMail(username)
	var indexData MailIndexData
	indexData.Comments = mail.comments.GetRootMail(username)
	indexData.Users = mail.users.GetUsers()

	/*
		tmpl, err := template.ParseFiles("templates/header.html", "templates/mailIndex.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//err = tmpl.ExecuteTemplate(w, "mailIndex.html", currentComments)
		err = tmpl.ExecuteTemplate(w, "mailIndex.html", indexData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/
	tpl := template.Must(template.ParseFiles("templates/mailIndex.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", indexData)
}

func (mail *MailHandler) IDHandler(w http.ResponseWriter, r *http.Request) {

	parent := r.PathValue("id")
	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	log.Println("In mailHandler looking for Parent:", parent)

	//currentComments := getChildComments(database, parent)

	//test child comment logic.
	//currentComments := mail.comments.GetMailComments(parent, username)

	var indexData MailIndexData
	indexData.Comments = mail.comments.GetMailComments(parent, username)
	indexData.User = mail.users.GetUser(username)

	tmpl := template.Must(template.ParseFiles("templates/mail.html"))
	tmpl.Execute(w, indexData)

}

func (mail *MailHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	username := ""
	id := ""
	root := ""
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
	mail.comments.EditCommentPic(id, username+"/"+filename)

	currentComments := mail.comments.GetMailComments(root, username)

	var indexData MailIndexData
	indexData.Comments = currentComments

	tmpl := template.Must(template.ParseFiles("templates/mail.html"))
	tmpl.ExecuteTemplate(w, "comment-list-element", indexData)

}
