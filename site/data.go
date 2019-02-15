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
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"net/http"
	"os"
)


func getBQData(r *http.Request) error {
	/*Trying out BigQuery*/
	ctx := context.Background()

	datasetID := os.Getenv("DATASETID")
	projectID := os.Getenv("PROJECTID")
	if datasetID == "" || projectID == ""{
		log.Print("IDs not set")
	}

	client, err := bigquery.NewClient(ctx, projectID);
	if err != nil {
		return err
	}

	q := client.Query(
		`SELECT *
    	FROM ` + "`flipthescript.fts.Media2`" + `
    	LIMIT 20`)
	// Location must match that of the dataset(s) referenced in the query.
	q.Location = "US"

	it, err := q.Read(ctx)
	if err != nil {
		return err
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("Title: %v and type %v\n", row[1], row[2])
	}
	/*it:= client.Datasets(ctx)*/
	/*for {
		dataset, err := it.Next()
		if err == iterator.Done {
			break
		}
		fmt.Println("Dataset id:", dataset.DatasetID)
	}
*/
	/*	job, err := q.Run(ctx)
		if err != nil {
			return err
		}
		status, err := job.Wait(ctx)
		if err != nil {
			return err
		}
		if err := status.Err(); err != nil {
			return err
		}
	*/
	return nil
}
