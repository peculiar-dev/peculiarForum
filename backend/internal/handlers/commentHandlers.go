package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"peculiarity/internal/data"
	"time"

	"github.com/google/uuid"
)

type CommentHandler struct {
	comments      data.Commentdb
	notifications data.Notificationdb
	users         data.Userdb
}

type CommentsData struct {
	Comments *[]data.Comment
	User     *data.User
}

func NewCommentHandler(commentdb data.Commentdb, notificationdb data.Notificationdb, userdb data.Userdb) *CommentHandler {
	return &CommentHandler{comments: commentdb, notifications: notificationdb, users: userdb}
}

func (ch *CommentHandler) IDHandler(w http.ResponseWriter, r *http.Request) {

	parent := r.PathValue("id")
	username := r.Header.Get("X-User")

	if username == "" {
		username = "test"
	}

	log.Println("In getCommentHandler looking for Parent:", parent)

	//currentComments := getChildComments(database, parent)

	//test child comment logic.
	//currentComments := ch.comments.GetChildComments(parent, username)

	var data CommentsData
	data.Comments = ch.comments.GetChildComments(parent, username)
	data.User = ch.users.GetUser(username)

	tmpl := template.Must(template.ParseFiles("templates/collapse.html"))
	tmpl.Execute(w, data)

}

func (ch *CommentHandler) AddHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	id := uuid.New().String()
	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	replyTo := r.FormValue("replyTo")
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

	ch.comments.InsertComment(id, r.FormValue("root"), username, message, linkAddr, parent, bRoot, bSticky)
	/* db.InsertNotification(Notification{Sender: "test2", Reciever: "test", CommentLink: "/comment/id-1/id-2", Created: time.Now()})
	 */

	rootComment := ch.comments.GetComment(r.FormValue("root"))

	ch.comments.EditComment(rootComment.Id, rootComment.Message, rootComment.Link, "root", true, rootComment.Sticky, time.Now())

	link := "/comment/" + r.FormValue("root") + "/" + id
	ch.notifications.InsertNotification(data.Notification{Sender: username, Reciever: replyTo, CommentLink: link, Created: time.Now()})
	/*
		AddCommentToSublist(&comments, parent, comment)
		log.Println("added to sublist")
	*/

	log.Println("printing comments from:", r.FormValue("root"))
	currentComments := ch.comments.GetChildComments(r.FormValue("root"), username)

	var data CommentsData
	data.Comments = currentComments

	//PrintComments(currentComments, "")

	tpl := template.Must(template.ParseFiles("templates/collapse.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", data)
}

func (ch *CommentHandler) EditHandler(w http.ResponseWriter, r *http.Request) {

	var bRoot bool
	var bSticky bool

	username := r.Header.Get("X-User")
	message := r.FormValue("comment")
	parent := r.FormValue("parent")
	id := r.FormValue("id")
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

	ch.comments.EditComment(id, message, linkAddr, parent, bRoot, bSticky, ch.comments.GetComment(id).Created)
	/*
		AddCommentToSublist(&comments, parent, comment)
		log.Println("added to sublist")
	*/

	log.Println("printing comments from:", r.FormValue("root"))
	currentComments := ch.comments.GetChildComments(r.FormValue("root"), username)

	var data CommentsData
	data.Comments = currentComments

	// PrintComments(currentComments, "")

	tpl := template.Must(template.ParseFiles("templates/collapse.html"))

	tpl.ExecuteTemplate(w, "comment-list-element", data)
}

func (ch *CommentHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	id := ""
	root := ""
	filename := ""

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
			// Process hidden fields (assuming they have a name)
			/*
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
			*/
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
	ch.comments.EditCommentPic(id, username+"/"+filename)

	currentComments := ch.comments.GetChildComments(root, username)

	var data CommentsData
	data.Comments = currentComments

	tpl := template.Must(template.ParseFiles("templates/collapse.html"))
	tpl.ExecuteTemplate(w, "comment-list-element", data)

}
