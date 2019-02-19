// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type RadioButton struct {
	Name       string
	Value      string
	IsDisabled bool
	IsChecked  bool
	Text       string
}

type templateParams struct {
	PageTitle       string
	Date 			string
	PageRadioButtons []RadioButton
	Answer          string

}

func main() {
	//Start the web server, set the port to listen to 8080. Without assumes localhost.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Print("Defaulting to port %s ", port)
	}
	registerHandlers()
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

//Code adjust from https://github.com/campoy/go-web-workshop/blob/master/section02/README.md & https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/app.go
func registerHandlers() {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").Handler(appHandler(indexHandler))
	r.Methods("POST").Path("/selected").Handler(appHandler(optionSelected))

	/*	r.Methods("POST").Path("/books").HandlerFunc(appHandler(createHandler))
	// match only POST requests on /media/
	r.Methods("POST").Handler("/media/", addMedia)
	// match GET regardless of mediaID
	r.HandleFunc("/media/{mediaID}", getMedia)
*/
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
}


// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v\n", e.Code, e.Message, e.Error)
		http.Error(w, e.Message, e.Code)
	}

}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}


func indexHandler(w http.ResponseWriter, r *http.Request) *appError{
	params := templateParams {}
	params.PageTitle = "Flip th Script"
	params.Date = time.Now().Format("02-01-2006")
	params.PageRadioButtons = []RadioButton{
		RadioButton{"animalselect", "cats", false, false, "Cats"},
		RadioButton{"animalselect", "dogs", false, false, "Dogs"},}


/*	t := template.Must(template.ParseFiles("content/index.html"))
*//*	if err := getBQData(r); err != nil {
		log.Fatalf("Error getting data: %v\n", err)
	}
*/
/*	//execute the template and pass it the params struct to fill in the gaps
	if err := t.ExecuteTemplate(w, "index.html", params); err !=nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return appErrorf(err, "Form execute error: %v", err)
	}*/
	log.Printf("INDEXHANDLER")
	return parseTemplate("index.html").Execute(w, r, params)
}

/*https://blog.scottlogic.com/2017/02/28/building-a-web-app-with-go.html*/
func optionSelected(w http.ResponseWriter, r *http.Request) *appError {
	r.ParseForm()
	// r.Form is now either
	// map[animalselect:[cats]] OR
	// map[animalselect:[dogs]]
	// so get the animal which has been selected
	log.Printf("OPTIONS %s", r.FormValue("animalselect"))
	params := templateParams {}
	params.PageTitle = "Your preferred animal"
	params.Date = time.Now().Format("02-01-2006")
	params.Answer = r.Form.Get("animalselect")


	// generate page by passing page variables into template
/*
	t := template.Must(template.ParseFiles("content/index.html"))

	//execute the template and pass it the params struct to fill in the gaps
	if err := t.ExecuteTemplate(w, "index.html", params); err !=nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return appErrorf(err, "Post execute error: %v", err)
	}*/
	log.Printf("OPTIONS")

	return parseTemplate("index.html").Execute(w,r,params)
}


// parseTemplate applies a given file to the body of the base template.
func parseTemplate(filename string) *appTemplate {
/*	tmpl := template.Must(template.ParseFiles("templates/base.html"))

	// Put the named file into a template called "body"
	path := filepath.Join("content", filename)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("could not read template: %v", err))
	}
	template.Must(tmpl.New("body").Parse(string(b)))
*/
	log.Printf("PARSING")
	tempPath :="content/"
	tmpl := template.Must(template.ParseFiles(tempPath + filename))

	//execute the template and pass it the params struct to fill in the gaps

	return &appTemplate{tmpl.Lookup("index.html")}
}

// appTemplate is a user login-aware wrapper for a html/template.
type appTemplate struct {
	t *template.Template
}

// Execute writes the template using the provided data, adding login and user
// information to the base template.
func (tmpl *appTemplate) Execute(w http.ResponseWriter, r *http.Request, data interface{}) *appError {
	d := struct {
		Data        interface{}
	}{
		Data:        data,
	}
	log.Printf("EXECUTE")
	if err := tmpl.t.Execute(w, d); err != nil {
		return appErrorf(err, "could not write template: %v", err)
	}
	return nil
}