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
	"cloud.google.com/go/bigquery"
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"os"
)

var mediaSchemaOrig = bigquery.Schema{
	{Name: "ID", Required: true, Type: bigquery.RecordFieldType},
	{Name: "Name", Required: true, Type: bigquery.StringFieldType},
	{Name: "MediaType", Required: true, Type: bigquery.StringFieldType},
	{Name: "DirectorName", Required: true, Type: bigquery.StringFieldType},
	{Name: "Industry", Required: true, Type: bigquery.StringFieldType},
}

var mediaSchema = bigquery.Schema{
	{Name: "ID", Required: true, Type: bigquery.RecordFieldType},
	{Name: "Name", Required: true, Type: bigquery.StringFieldType},
	{Name: "ReleaseDate", Repeated: true, Type: bigquery.DateFieldType},
	{Name: "MediaType", Required: true, Type: bigquery.StringFieldType},
	{Name: "Rating", Repeated: true, Type: bigquery.IntegerFieldType},
	{Name: "BechdelPass", Required: false, Type: bigquery.BooleanFieldType},
	{Name: "WikiURL", Required: false, Type: bigquery.StringFieldType},
	{Name: "IMDBLink", Required: false, Type: bigquery.StringFieldType},
	{Name: "RottenTomatoeLink", Required: false, Type: bigquery.StringFieldType},
}


var actorSchema = bigquery.Schema{
	{Name: "Name", Required: true, Type: bigquery.StringFieldType},
	{Name: "ReleaseDate", Repeated: true, Type: bigquery.DateFieldType},
	{Name: "MediaType", Required: true, Type: bigquery.StringFieldType},
	{Name: "Rating", Repeated: true, Type: bigquery.IntegerFieldType},
	{Name: "BechdelPass", Required: false, Type: bigquery.BooleanFieldType},
	{Name: "WikiURL", Required: false, Type: bigquery.StringFieldType},
	{Name: "IMDBLink", Required: false, Type: bigquery.StringFieldType},
	{Name: "RottenTomatoeLink", Required: false, Type: bigquery.StringFieldType},
}

func setupQuery(d string) (string){

}

/*BigQuery query*/
func getBQData(q string) ([][]bigquery.Value, error) {

	ctx := context.Background()

	datasetID := os.Getenv("DATASETID")
	projectID := os.Getenv("PROJECTID")
	//projectID := appengine.AppID(ctx) // should work with appengine
	if datasetID == "" || projectID == ""{
		log.Print("SETUP ENVIRONMENT VARIABLES")
	}

	client, err := bigquery.NewClient(ctx, projectID);
	if err != nil {
		return nil, err

	}

	query := client.Query(q)

	// Location must match that of the dataset(s) referenced in the query.
	query.Location = "US"

	it, err := query.Read(ctx)
	if err != nil {
		return nil, err

	}

	var rows [][]bigquery.Value
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			return rows, nil
		}
		if err != nil {
			return nil, err

		}
		/*fmt.Printf("Title: %v and type %v\n", row[1], row[2])*/
		rows = append(rows, row)
	}
}


type mediaParams struct {
	MediaID       	bigquery.Value
	Title 			bigquery.Value
	MediaType 		bigquery.Value
	DirectorName    bigquery.Value
	Industry	    bigquery.Value
	ImageURL		bigquery.Value
	ReleaseDate		bigquery.Value
	Description		bigquery.Value
}


func listMedia(query string) ([]mediaParams, error) {
	mediaList, err := getBQData(query)
	if err != nil {
		return nil, err
	}

	var mediaStruct []mediaParams
	for _, row := range mediaList {
		/*add to website*/
		if debugProject { fmt.Sprintf("%s",row) }
		media := mediaParams{}
		media.MediaID =  row[0]
		media.Title = row[1]
		media.MediaType = row[2]
		media.DirectorName = row[3]
		media.Industry = row[4]
		mediaStruct = append(mediaStruct, media)
	}
	return mediaStruct, nil
}

// GetBook retrieves a media by its ID.
func () GetMedia(id int64) ([]mediaParams, error) {
	query := setupQuery(id)
	query := `SELECT ID + ' id 
    	FROM ` + "`flipthescript.fts.Media`" + `
    	LIMIT 5`

	media, err := listMedia(query)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Could not find media with id %d", id)
	}

	if err != nil {
		return nil, fmt.Errorf("Could not get specific media: %v", err)
	}
		return media, nil
}