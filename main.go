package main

import (
	"html/template"
	"net/http"
	"os"

	// using https://github.com/apex/up for Amazon Web Services (AWS)
	// AWS Lambda is a simple, cost effective solution to host such service
	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"

	// using this open source software to manipulate the template
	"github.com/nguyenthenguyen/docx"
)

// The html form lives here
var views = template.Must(template.ParseGlob("views/*.html"))

// use JSON logging when run by Up (including `up start`).
func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(json.Default)
	}
}

// setup.
func main() {
	addr := ":" + os.Getenv("PORT")

	// when submit button on the html is pressed the submit function (line 61)  is invoked
	http.HandleFunc("/submit", submit)

	// this handles the index or landing page currently the form
	http.HandleFunc("/", index)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error listening: %s", err)
	}
}

// index page.
func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	views.ExecuteTemplate(w, "index.html", struct {
		Name  string
		Email string
	}{})
}

// submit handler.
func submit(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")

	log.WithFields(log.Fields{
		"name":  name,
		"email": email,
	}).Info("submit")

	doc, err := docx.ReadDocxFile("views/Form1.docx")
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer doc.Close()

	docx1 := doc.Editable()

	// This is where we start replacing the template document with data from the html form
	// the word document template has place holders like %name% that map to the HTML form "name"
	docx1.Replace("%name%", name, -1)
	docx1.Replace("%email%", email, -1)

	// this sets the correct header for a microsoft word document download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	// this names the downloaded file to form1.docx
	w.Header().Set("Content-Disposition", "attachment; filename=form1.docx")

	// this will make the download start
	err = docx1.Write(w)

	// if we error out in starting the download
	if err != nil {
		// we log the error for debugging
		log.Errorf("Error writing file: %v", err)
	}

}

// redirect to referrer helper.
func redirectBack(w http.ResponseWriter, r *http.Request) {
	url := r.Header.Get("Referer")
	http.Redirect(w, r, url, http.StatusFound)
}

// cookie helper.
func cookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return c.Value
}
