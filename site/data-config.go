// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"os"

	"cloud.google.com/go/datastore"

	"cloud.google.com/go/storage"
	"github.com/gorilla/sessions"
)

var (

	StorageBucket		*storage.BucketHandle
	StorageBucketName	string

	SessionStore		sessions.Store
	PubsubClient 	*pubsub.Client


)

const PubsubTopicID = "fill-media-details"

/*func init() {
	var err error
*/
// To use the in-memory test database, uncomment the next line.
/*DB = newMemoryDB()*/

// [START cloudsql]
// To use Cloud SQL, uncomment the following lines, and update the username,
// password and instance connection string. When running locally,
// localhost:3306 is used, and the instance name is ignored.
/*
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: os.Getenv("SQL_USERNAME"),
		Password: os.Getenv("SQL_PWD"),
		Instance: os.Getenv("SQL_INSTANCE"),
	})
*/
// 	// The connection name of the Cloud SQL v2 instance, i.e.,
// 	// "project:region:instance-id"
// 	// Cloud SQL v1 instances are not supported.
// 	Instance: "",
// })
// [END cloudsql]
/*
	if err != nil {
		log.Fatal(err)
	}*/

// [START datastore]
// To use Cloud Datastore, uncomment the following lines and update the
// project ID.
// More options can be set, see the google package docs for details:
// http://godoc.org/golang.org/x/oauth2/google
//
// DB, err = configureDatastoreDB("<your-project-id>")
// [END datastore]

/*	if err != nil {
		log.Fatal(err)
	}
*/

// [START storage]
// To configure Cloud Storage, uncomment the following lines and update the
// bucket name.
//
// StorageBucketName = "<your-storage-bucket>"
// StorageBucket, err = configureStorage(StorageBucketName)
// [END storage]

/*	if err != nil {
		log.Fatal(err)
	}
*/
// [START pubsub]
// To configure Pub/Sub, uncomment the following lines and update the project ID.
//
// PubsubClient, err = configurePubsub("<your-project-id>")
// [END pubsub]

/*	if err != nil {
		log.Fatal(err)
	}*/
/*}
*/

func configureDatastoreDB(projectID string) (MediaDatabase, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return newDatastoreDB(client)
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}


type cloudSQLConfig struct {
	Username, Password, Instance, IP string
	PgPort                       int
}

func configureCloudSQL(config cloudSQLConfig) (MediaDatabase, error) {
	if os.Getenv("GAE_INSTANCE") != "" {
		// Running in production.
		return newPgSQLDB(PgSQLConfig{
			Username:   config.Username,
			Password:   config.Password,
			UnixSocket:   "/cloudsql/" + config.Instance,
			Instance:   config.Instance,
			IP: 		config.IP,
		})
	}

	// Running locally.
	return newPgSQLDB(PgSQLConfig{
		Username: config.Username,
		Password: config.Password,
		Host:     "localhost",
		Port:     config.PgPort,
	})
}