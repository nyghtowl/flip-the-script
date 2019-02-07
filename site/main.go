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
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"time"
	"html/template"
)

type PageVariables struct {
	Name string
	Date string
	Time string
}

func main() {
	//Start the web server, set the port to listen to 8080. Without assumes localhost.
	port := os.Getenv("PORT")
	if port == ""{
		port = "8080"
		log.Print("Defaulting to port %s ", port)
	}

	//Code adjust from https://github.com/campoy/go-web-workshop/blob/master/section02/README.md
	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)
	http.Handle("/", r)

	//// match only GET requests on /media/
	//r.HandleFunc("/media/", listMedia).Methods("GET")
	//
	//// match only POST requests on /media/
	//r.HandleFunc("/media/", addMedia).Methods("POST")
	//
	//// match GET regardless of mediaID
	//r.HandleFunc("/media/{mediaID}", getMedia)


	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	now := time.Now() // find the time right now
	HomePageVars := PageVariables{ //store the date and time in a struct
		Name: "Flip the Script",
		Date: now.Format("02-01-2006"),
		Time: now.Format("15:04:05"),
	}

	t := template.Must(template.ParseFiles("content/index.html")) //parse the html file index.html
	//if err != nil { // if there is an error
	//	log.Print("template parsing error: ", err) // log it
	//}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	 //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err := t.ExecuteTemplate(w, "index.html", HomePageVars); err !=nil { // if there is an error
		http.Error(w, err.Error(), http.StatusInternalServerError) }//log it


}
