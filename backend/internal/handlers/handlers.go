package handlers

import (
	"automate/internal/fileUtils"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Test struct {
	Name       string
	ID         string
	User       string
	LogLink    string
	LogImg     string
	ResultLink string
	ResultImg  string
	QueueTime  string
	EndTime    string
	State      string
}

type TestData struct {
	Tests []Test
}

var Directory string
var fileRegex *regexp.Regexp
var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// grab the request.MultipartReader
	log.Println("uploading file")

	reader, err := r.MultipartReader()
	if err != nil {
		log.Println("multipart error:" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		// if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			log.Println("filename empty")
			http.Error(w, "filename empty", http.StatusBadRequest)
			return
		}

		goodFile := fileRegex.Match([]byte(part.FileName()))

		if !goodFile || err != nil {
			log.Println("filename incorrect, should follow format ID-test.prc. Exampe: 7f2150622ad54e839677a916c745d20f-gms20-nettest.spec.js.prc")
			http.Error(w, "filename incorrect, should follow format ID-test.prc. Exampe: 7f2150622ad54e839677a916c745d20f-gms20-nettest.spec.js.prc", http.StatusBadRequest)
			return
		}

		// create the destination file
		log.Println("creating file:", Directory+part.FileName())
		dst, err := os.Create(Directory + part.FileName())
		fileUtils.PushLine(part.FileName()+"\n", Directory+"queue.txt")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// copy the part to dst
		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func LogHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	filename := Directory + id + "-out.log"
	log.Println("looking for file:" + filename)

	tpl, err := template.ParseFiles(filename)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		tpl.ExecuteTemplate(w, id+"-out.log", nil)
	}

}

func ResultHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	filename := Directory + id + ".html"
	log.Println("looking for file:" + filename)

	tpl, err := template.ParseFiles(filename)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		tpl.ExecuteTemplate(w, id+".html", nil)
	}

}

func testData(search string) *TestData {

	f, err := os.Open(Directory)
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
	// ignore folders and files that don't contain search string
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), search) && strings.Contains(file.Name(), ".prc") {
			onlyfiles = append(onlyfiles, file)
		}
	}
	sort.Slice(onlyfiles, func(i, j int) bool {
		return onlyfiles[j].ModTime().Before(onlyfiles[i].ModTime())
	})
	tests := make([]Test, 0)
	for _, file := range onlyfiles {

		test := new(Test)

		args := strings.Split(file.Name(), "-")

		test.ID = args[0]
		test.User = args[1]
		test.Name = args[2][:len(args[2])-4]
		test.QueueTime = file.ModTime().Format("2006-01-02 15:04:05")
		test.LogImg = "/static/comingsoon.gif"
		test.ResultImg = "/static/comingsoon.gif"

		// check to see if the log file exists
		if _, err := os.Stat(Directory + test.ID + "-out.log"); errors.Is(err, os.ErrNotExist) {
			test.State = "queued"
		} else {
			test.State = "running"

			test.LogImg = "/static/newblast.gif"
			test.LogLink = "/log/" + test.ID
		}
		// check to see if the result HTML exists
		if resultFile, err := os.Stat(Directory + test.ID + ".html"); !errors.Is(err, os.ErrNotExist) {
			test.State = "complete"
			test.EndTime = resultFile.ModTime().Format("2006-01-02 15:04:05")
			test.ResultImg = "/static/newblast.gif"
			test.ResultLink = "/results/" + test.ID
		}

		tests = append(tests, *test)
	}

	testData := TestData{tests}
	return &testData

}

func TestSearchHandler(w http.ResponseWriter, r *http.Request) {

	searchString := r.PathValue("search")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testData(searchString))

}

func DashHandler(w http.ResponseWriter, r *http.Request) {

	//enabled only for development. No need to reload templates in production!
	tpl = template.Must(template.ParseGlob("templates/*"))

	searchString := r.PathValue("search")

	err := tpl.ExecuteTemplate(w, "dashboard.html", testData(searchString))

	if err != nil {
		log.Println("Error processing Dashboard: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func init() {


	if r, err := regexp.Compile(`^[a-zA-Z0-9]+-[a-zA-Z0-9]+-[a-zA-Z0-9_]+(?:\.[a-zA-Z]+)*\.prc$`); err != nil {
		log.Fatal("error compiling regex")
	} else {
		fileRegex = r
	}
}
