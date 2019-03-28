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
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	// See data-scratch.go
	DB				MediaDatabase
	DBName = os.Getenv("DBNAME")
	TABLENAME = os.Getenv("TABLENAME")

	// See template.go
	indexTmpl   = parseTemplate("index.html")
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
	detailBQTmpl = parseTemplate("detailBQ.html")

	debugProject = true
	bigQueryClient *bigquery.Client
	projectID, datasetID string
)

/*
TODO break down the index file into components - add list section and pass in data - remove hardcode of page in TemplatesTODO var for all the things esp images and db names
TODO all form input - get it
TODO add tests
TODO add actors, characters, directors and expand on media
TODO add in memory, datastore and pubsub to store data loaded
TODO put flipthescript domain in place and upload on AE
*/


func init() {
	var err error
	projectID = os.Getenv("PROJECTID")
	datasetID = os.Getenv("DATASETID")
	if datasetID == "" || projectID == ""{
		log.Print("SETUP ENVIRONMENT VARIABLES")
	}
	bigQueryClient, err = bigquery.NewClient(context.Background(), projectID);
	if err != nil {
		log.Fatalf("Cannot initialize BigQuery client: %v, ", err)
	}
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: os.Getenv("PgSQL_USERNAME"),
		Password: os.Getenv("PgSQL_PWD"),
		Instance: os.Getenv("PgSQL_INSTANCE"),
		IP: os.Getenv("PgSQL_IP"),
	})
	// 	// The connection name of the Cloud SQL v2 instance, i.e.,
	// 	// "project:region:instance-id"
	// 	// Cloud SQL v1 instances are not supported.
	// 	Instance: "",
	// })
	if err != nil {
		log.Fatal(err)
	}
}

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
	/*r.Methods("GET").Path("/media").
		Handler(appHandler(listHandlerBQ))*/
/*	r.Methods("GET").Path("/media/{id:[0-9]+}").
		Handler(appHandler(detailHandlerBQ))
*/	r.Methods("GET").Path("/media").Handler(appHandler(indexHandler))
	r.Methods("GET").Path("/media/").Handler(appHandler(indexHandler))
	r.Methods("GET").Path("/media/list").Handler(appHandler(listHandler))
	r.Methods("GET").Path("/media/{id:[0-9]+}").
		Handler(appHandler(detailHandler))
	r.Methods("GET").Path("/media/add").
		Handler(appHandler(addFormHandler))
	r.Methods("GET").Path("/media/{id:[0-9]+}/edit").
		Handler(appHandler(editFormHandler))

	r.Methods("POST").Path("/media").
		Handler(appHandler(createHandler))
	r.Methods("POST", "PUT").Path("/media/{id:[0-9]+}").
		Handler(appHandler(updateHandler))
	r.Methods("POST").Path("/media/{id:[0-9]+}:delete").
		Handler(appHandler(deleteHandler)).Name("delete")

	// Respond to App Engine and Compute Engine health checks.
	// Indicate the server is healthy.
	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

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

/*---------------------------  BigQuery  ---------------------------*/


// listHandler displays a list with summaries media in the database.
func listHandlerBQ(w http.ResponseWriter, r *http.Request) error {

	/*TODO Create query - setup to it can be generalized*/
	defaultQuery := `SELECT *
    	FROM ` + "`flipthescript.fts.Media`" + `
    	LIMIT 20`

	media, err := listMedia(r.Context(), defaultQuery)
	if err != nil {
		return appErrorf(err, "Error getting data from BigQuery: %v", err)
	}

	return listTmpl.Execute(w, r, media)
}

func mediaFromBQRequest(r *http.Request) ([]MediaBQ, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad media id: %v", err)
	}
	media, err := GetMedia(r.Context(), id)
	if err != nil {
		return nil, fmt.Errorf("could not find media: %v", err)
	}
	return media, nil
}

// detailHandler displays the details of given media.
func detailHandlerBQ(w http.ResponseWriter, r *http.Request) error {
	media, err := mediaFromBQRequest(r)
	log.Printf("DETAILED HANDLER %v", media)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return detailBQTmpl.Execute(w, r, media)
}

/*---------------------------  Cloud SQL  ---------------------------*/

//index is the start page
func indexHandler(w http.ResponseWriter, r *http.Request) error {
	log.Printf("INDEX HANDLER")
	return indexTmpl.Execute(w, r, nil)
}

// listHandler displays a list with summaries media in the database.
func listHandler(w http.ResponseWriter, r *http.Request) error {
	log.Printf("LIST HANDLER")
	media, err := DB.ListMedia()
	if err != nil {
		return appErrorf(err, "could not list media: %v", err)
	}
	media[0].PageSubTitle = "Media List"
	return listTmpl.Execute(w, r, media)
}

// bookFromRequest retrieves media from the database given a media ID in the
// URL's path.
func mediaFromRequest(r *http.Request) (*Media, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad media id: %v", err)
	}
	media, err := DB.GetMedia(id)
	if err != nil {
		return nil, fmt.Errorf("could not find media: %v", err)
	}
	return media, nil
}

// detailHandler displays the details of a given an item.
func detailHandler(w http.ResponseWriter, r *http.Request) error {
	media, err := mediaFromRequest(r)
	if err != nil {
		return appErrorf(err, "could not list media detail: %v", err)
	}
	media.PageSubTitle = "Media Details"
	return detailTmpl.Execute(w, r, media)
}


// addFormHandler displays a form that captures details of a new item to add to
// the database.
func addFormHandler(w http.ResponseWriter, r *http.Request) error {
	return editTmpl.Execute(w, r, nil)
}

// editFormHandler displays a form that allows the user to edit the details of
// a given item.
func editFormHandler(w http.ResponseWriter, r *http.Request) error {
	media, err := mediaFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	media.PageSubTitle = "Edit Media"
	return editTmpl.Execute(w, r, media)
}

// mediaFromForm populates the fields of a Book from form values
// (see templates/edit.html).
func mediaFromForm(r *http.Request) (*Media, error) {
	/*TODO confirm exists or doesn't*/
	log.Printf(" MEDIA FROM FORM")

	/*imageURL, err := uploadFileFromForm(r) TODO store form & image and return image link
	if err != nil {
		return nil, fmt.Errorf("could not upload file: %v", err)
	}
	if imageURL == "" {
		imageURL = r.FormValue("imageURL")
	}*/

/*	actorID, _ := strconv.ParseInt(r.FormValue("actorID"), 10, 64)
	characterID, _ := strconv.ParseInt(r.FormValue("characterID"), 10, 64)
	dirctorID, _ := strconv.ParseInt(r.FormValue("dirctorID"), 10, 64)
	bechdel, _ := strconv.ParseBool(r.FormValue("bechdel"))
*/
	actorID, _ := strconv.ParseInt(strconv.Itoa(rand.Intn(30)), 10, 64)
	characterID, _ := strconv.ParseInt(strconv.Itoa(rand.Intn(30)), 10, 64)
	dirctorID, _ := strconv.ParseInt(strconv.Itoa(rand.Intn(30)), 10, 64)
	bechdel := false

		media := &Media{
		Title:         r.FormValue("title"),
		Description:   r.FormValue("description"),
		MediaType: 	   r.FormValue("mediaType"),
		Industry:	   r.FormValue("industry"),
		ReleaseDate:   r.FormValue("releaseDate"),

		ActorID:	   actorID,
		CharacterID:   characterID,
		DirectorID:    dirctorID,

		ImageURL:      r.FormValue("imageURL"),

		Bechdel:	   bechdel,
		WikiURL:	   r.FormValue("imageURL"),
		IMDBURL:	   r.FormValue("imageURL"),
		RottenTomURL:  r.FormValue("imageURL"),

		CreatedBy:     r.FormValue("createdBy"),
	}

	log.Printf(" MEDIA | %v", media)

	// Anonymous if the createdBy is empty
	if media.CreatedBy == "" {
		media.SetCreatorAnonymous()
	} else {
		media.CreatedByID, _ = strconv.ParseInt(strconv.Itoa(rand.Intn(30)), 10, 64)
		//Assume if empty z
		media.CreatedDate = time.Now().Format("02-01-2006")
	}

	return media, nil
}


// createHandler adds a media to the database.
func createHandler(w http.ResponseWriter, r *http.Request) error {
	media, err := mediaFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse media from form: %v", err)
	}
	id, err := DB.AddMedia(media)
	if err != nil {
		return appErrorf(err, "could not save media: %v", err)
	}
/*	TODO go publishUpdate(id)
*/	http.Redirect(w, r, fmt.Sprintf("/media/%d", id), http.StatusFound)
	return nil
}

// updateHandler updates the details of a given book.
func updateHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad media id: %v", err)
	}

	media, err := mediaFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse media from form: %v", err)
	}
	media.ID = id

	err = DB.UpdateMedia(media)
	if err != nil {
		return appErrorf(err, "could not save media: %v", err)
	}
/*	TODO go publishUpdate(media.ID)
*/
	media.PageSubTitle = "Media Update"
	http.Redirect(w, r, fmt.Sprintf("/media/%d", media.ID), http.StatusFound)
	return nil
}

// deleteHandler deletes a given book.
func deleteHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad media id: %v", err)
	}
	err = DB.DeleteMedia(id)
	if err != nil {
		return appErrorf(err, "could not delete media: %v", err)
	}
	http.Redirect(w, r, "/media", http.StatusFound)
	return nil
}

// publishUpdate notifies Pub/Sub subscribers that the media identified with
// the given ID has been added/modified.
func publishUpdate(mediaID int64) {
	if PubsubClient == nil {
		return
	}

	ctx := context.Background()

	m, err := json.Marshal(mediaID)
	if err != nil {
		return
	}
	topic := PubsubClient.Topic(PubsubTopicID)
	_, err = topic.Publish(ctx, &pubsub.Message{Data: m}).Get(ctx)
	log.Printf("Published update to Pub/Sub for Media ID %d: %v", mediaID, err)
}



