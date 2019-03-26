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
	"strconv"
	_ "strconv"
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


/*BigQuery query*/
func getBQData(ctx context.Context, q string) ([][]bigquery.Value, error) {

	query := bigQueryClient.Query(q)

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

type MediaBQ struct {
	ID		       	bigquery.Value
	Title 			bigquery.Value
	Description		bigquery.Value
	MediaType 		bigquery.Value
	Industry	    bigquery.Value
	ReleaseDate	    bigquery.Value

	ActorID		 	bigquery.Value
	CharacterID	 	bigquery.Value
	DirectorID	    bigquery.Value

	ImageURL		bigquery.Value
	Bechdel		    bigquery.Value
	WikiURL		    bigquery.Value
	IMDBURL		    bigquery.Value
	RottenTomURL    bigquery.Value

	CreatedByID	  	bigquery.Value
	CreatedBy		bigquery.Value
	CreatedDate	    bigquery.Value
}

func listMedia(ctx context.Context, query string) ([]MediaBQ, error) {
	mediaList, err := getBQData(ctx, query)
	if err != nil {
		return nil, err
	}

	var mediaStruct []MediaBQ
	for _, row := range mediaList {
		/*add to website*/
		if debugProject { fmt.Sprintf("%s",row) }
		media := MediaBQ{}
		media.ID =  row[0]
		media.Title = row[1]
		media.MediaType = row[2]
		media.DirectorID = row[3]
		media.Industry = row[4]
		mediaStruct = append(mediaStruct, media)
	}
	return mediaStruct, nil
}

// GetBook retrieves a media by its ID.
func GetMedia(ctx context.Context, id int64) ([]MediaBQ, error) {
	query := `SELECT * FROM` + "`flipthescript.fts.Media`" + `
    	WHERE ID=` + strconv.FormatInt(id, 10)

	media, err := listMedia(ctx, query)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Could not find media with id %d", id)
	}

	if err != nil {
		return nil, fmt.Errorf("Could not get specific media: %v", err)
	}
		return media, nil
}