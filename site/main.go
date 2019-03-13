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
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	// See template.go
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
)

/*
TODO try other templates
TODO all form input
TODO put flipthescript domain in place

*/

var debugProject = true

func main() {
	//Start the web server, set the port to listen to 8080. Without assumes localhost.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s ", port)
	}
	registerHandlers()
	fmt.Sprintf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

//Code adjust from https://github.com/campoy/go-web-workshop/blob/master/section02/README.md & https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/app.go
func registerHandlers() {
	r := mux.NewRouter()

	/*Static file management*/
	staticHandler := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	/*Page routes*/
	r.Handle("/", http.RedirectHandler("/media", http.StatusFound))
	r.Methods("GET").Path("/media").
		Handler(appHandler(listHandler))
	r.Methods("GET").Path("/media/{id:[0-9]+}").
		Handler(appHandler(detailHandler))

	// match only POST requests on /media/
/*	r.Methods("POST").Handler("/media/", addMedia)
*/
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
}


// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) error

type appError struct {
	cause   error
	message string
	code    int
}

func (e *appError) Error() string {
	return fmt.Sprintf("Handler error: status code: %d, message: %s, underlying err: %#v\n", e.code, e.message, e.cause)
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e := fn(w, r)
	if appErr,ok := e.(*appError); ok  { // e is *appError, not os.Error.
		http.Error(w, appErr.message, appErr.code)
	}

}

func appErrorf(err error, format string, v ...interface{}) error {
	return &appError{
		cause:   err,
		message: fmt.Sprintf(format, v...),
		code:    500,
	}
}

// listHandler displays a list with summaries media in the database.
func listHandler(w http.ResponseWriter, r *http.Request) error {

	/*TODO Create query - setup to it can be generalized*/
	defaultQuery := `SELECT *
    	FROM ` + "`flipthescript.fts.Media`" + `
    	LIMIT 20`

	media, err := listMedia(defaultQuery)
	if err != nil {
		return appErrorf(err, "Error getting data from BigQuery: %v", err)
	}

	return listTmpl.Execute(w, r, media)
}

/*TODO Need to pass reference to database instead of connecting to it everytime ...   *mediaDB */

func mediaFromRequest(r *http.Request) ([]Media, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad book id: %v", err)
	}
	media, err := GetMedia(id)
	if err != nil {
		return nil, fmt.Errorf("could not find book: %v", err)
	}
	return media, nil
}

// detailHandler displays the details of given media.
func detailHandler(w http.ResponseWriter, r *http.Request) error {
	media, err := mediaFromRequest(r)
	log.Printf("DETAILED HANDLER %v", media)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return detailTmpl.Execute(w, r, media)
}


