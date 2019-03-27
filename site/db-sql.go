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
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"math/rand"
	"time"

	"log"
)


/*---------------------------  Core Structures  ---------------------------*/

// pgsqlDB persists media to a PostgreSQL instance.
type pgsqlDB struct {
	conn *sql.DB

	list   *sql.Stmt
	listBy *sql.Stmt
	insert *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

type PgSQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	Instance string

	IP string
	// UnixSocket is the filepath to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string
}

// Ensure pgsqlDB conforms to the MediaDatabase interface.
var _ MediaDatabase = &pgsqlDB{}

var mySQLUsed = false

// dataStoreName returns a connection string suitable for sql.Open.
func (c PgSQLConfig) dataStoreName() string {
	var cred string
	// [username[:password]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}

	if c.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, c.UnixSocket, DBName)
	}
	if mySQLUsed {
		return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, DBName) }


	return fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		c.Instance, DBName, c.Username, c.Password)
}

/*---------------------------  Statements  ---------------------------*/

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS media DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE media;`,
	`CREATE TABLE IF NOT EXISTS media (
		id INT PRIMARY KEY,
		title VARCHAR(255) NULL,
		description TEXT NULL,
		mediaType VARCHAR(255) NULL,
		releaseDate VARCHAR(255) NULL,
		industry VARCHAR(255) NULL,
		actorId INT NULL,
		characterId INT NULL,
		directorId INT NULL,
		imageURL VARCHAR(255) NULL,
		bechdel VARCHAR(255) NULL,
		wikiURL VARCHAR(255) NULL,
		imdbURL VARCHAR(255) NULL,
		rottentomURL VARCHAR(255) NULL,
		createdById INT NULL,
		createdBy VARCHAR(255) NULL,
		createdDate VARCHAR(255) NULL
	)`,
}

const getStatement = `SELECT * FROM media WHERE id = $1`

const listStatement = `SELECT * FROM media ORDER BY title`

const listByStatement = `SELECT * FROM media WHERE createdbyid = $1 ORDER BY title`

const insertStatement = `
  INSERT INTO media (title, description, mediaType,
		industry, releaseDate, actorID, characterID,
		directorID, imageURL, bechdel, wikiURL, imdbURL,
		rottentomURL, createdByID, createdBy, createdDate
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

const deleteStatement = `DELETE FROM media WHERE id = $1`

const updateStatement = `
  UPDATE media
  SET title=$1, description=$2, mediaType=$3, industry=$4, 
  		releaseDate=$5, actorID=$6, characterID=$7, directorID=$8, 
  		imageURL=$9, bechdel=$10, wikiURL=$11, imdbURL=$12, 
  		rottentomURL=$13, createdById=$14, createdBy=$15, createdDate=$16
  WHERE id = $17`

/*---------------------------  Core Functions  ---------------------------*/

// newPgSQLDB creates a new MediaDatabase backed by a given PgSQL server.
func newPgSQLDB(config PgSQLConfig) (MediaDatabase, error) {

	/*conn, err := sql.Open("postgres", config.dataStoreName())*/

	/*TODO verify which to use and pass back to datastore name*/
	conn, err := sql.Open("postgres", fmt.Sprintf(`sslmode=require sslrootcert=/Users/warrick/homebrew/etc/openssl/certs/server-ca.pem sslcert=/Users/warrick/homebrew/etc/openssl/certs/client-cert.pem sslkey=/Users/warrick/homebrew/etc/openssl/certs/client-key.pem host=%s user=%s dbname=%s password=%s`, config.IP, config.Username, DBName, config.Password))


/*	conn, err := sql.Open("cloudsqlpostgres", fmt.Sprintf(`host=%s dbname=%s user=%s password=%s sslmode=verify-ca sslrootcert=/Users/warrick/homebrew/etc/openssl/certs/server-ca.pem sslcert=/Users/warrick/homebrew/etc/openssl/certs/client-cert.pem sslkey=/Users/warrick/homebrew/etc/openssl/certs/client-key.pem"`, config.Instance, DBName, config.Username, config.Password))
*/
	if err != nil {
		return nil, fmt.Errorf("postgreSQL: could not get a connection: %v", err)
	}

	// Check database and table exists. If not, create it.
	if err := config.ensureTableExists(conn); err != nil {
		return nil, err
	}

	db := &pgsqlDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addBook).
	if db.get, err = conn.Prepare(getStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare get: %v", err)
	}
	if db.list, err = conn.Prepare(listStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare list: %v", err)
	}
	if db.listBy, err = conn.Prepare(listByStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare listBy: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare insert: %v", err)
	}
	if db.update, err = conn.Prepare(updateStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare update: %v", err)
	}
	if db.delete, err = conn.Prepare(deleteStatement); err != nil {
		return nil, fmt.Errorf("postgreSQL: prepare delete: %v", err)
	}

	return db, nil
}

/*TODO Make sure this is checking if table exists - changed dataStoreName to pass name*/

func (config PgSQLConfig) ensureTableExists(conn *sql.DB) error {

	// Check the connection.
	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("postgreSQL: could not establish a good connection: %v", err)
	}

	/*TODO generalize media name to pass in*/

	if _, err := conn.Exec(`SELECT EXISTS (SELECT * FROM media)`); err != nil {
			// MySQL error 1146, PgSQL error 42P01 is "table does not exist"
			log.Printf("TABLE NOT EXIST: %v", err)
			// Unknown error.
			return fmt.Errorf("postgreSQL: could not connect to the database: %v", err)
		}

	log.Printf("TABLE EXISTs")

	return nil
}

// Close closes the database, freeing up any resources.
func (db *pgsqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanBook reads a book from a sql.Row or sql.Rows
func scanMedia(s rowScanner) (*Media, error) {
	var (
		id            int64
		title         sql.NullString
		description   sql.NullString
		mediaType	  sql.NullString
		industry	  sql.NullString
		releaseDate  sql.NullString

		actorID       sql.NullInt64
		characterID   sql.NullInt64
		directorID    sql.NullInt64

		imageURL      sql.NullString
		bechdel		  sql.NullBool
		wikiURL		  sql.NullString
		imdbURL		  sql.NullString
		rottentomURL  sql.NullString

		createdByID   sql.NullInt64
		createdBy     sql.NullString
		createdDate   sql.NullString
	)

	if err := s.Scan(&id, &title, &description, &mediaType,
		&industry, &releaseDate, &actorID, &characterID,
		&directorID, &imageURL, &bechdel, &wikiURL, &imdbURL,
		&rottentomURL, &createdByID, &createdBy, &createdDate); err != nil {
		return nil, err
	}

	media := &Media{
		ID:            id,
		Title:         title.String,
		Description:   description.String,
		MediaType: 	   mediaType.String,
		Industry:      industry.String,
		ReleaseDate:   releaseDate.String,

		ActorID:	   actorID.Int64,
		CharacterID:   characterID.Int64,
		DirectorID:    directorID.Int64,

		ImageURL:      imageURL.String,
		Bechdel:	   bechdel.Bool,
		WikiURL:	   wikiURL.String,
		IMDBURL:	   imdbURL.String,
		RottenTomURL:  rottentomURL.String,

		CreatedByID:   createdByID.Int64,
		CreatedBy:     createdBy.String,
		CreatedDate:   createdDate.String,

	}
	return media, nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("postgreSQL: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("postgreSQL: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("postgreSQL: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}


/*---------------------------  Get/List  ---------------------------*/

// GetMedia retrieves media by its ID.
func (db *pgsqlDB) GetMedia(id int64) (*Media, error) {
	media, err := scanMedia(db.get.QueryRow(id))

	log.Printf("GET MEDIA | %v", media)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("postgreSQL: could not find media with id %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("postgreSQL: could not get media: %v", err)
	}
	return media, nil
}

// ListMedia returns a list of media, ordered by title.
func (db *pgsqlDB) ListMedia() ([]*Media, error) {
	rand.Seed(time.Now().Unix())
	tempPhotoList := []string{
		"http://media.tmz.com/2013/12/13/1213-grumpy-1.jpg",
		"https://blog.hubspot.com/hubfs/cats-hollywood.jpg",
		"https://www.fuzzyfuzzlet.com/MTV/Dollface-Persian-Kitten_Directors-Chair.jpg",
		"https://placekitten.com/g/200/300",
	}

	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*Media
	n:=0
	for rows.Next() {
		media, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("postgreSQL: could not read row: %v", err)
		}
		if media.ImageURL == "" {
			n = rand.Int() % len(tempPhotoList)
		}
		media.ImageURL = (tempPhotoList[n])
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// ListBooksCreatedBy returns a list of books, ordered by title, filtered by
// the user who created the book entry.
func (db *pgsqlDB) ListMediaCreatedBy(userID int64) ([]*Media, error) {
	if userID == 0 {
		return db.ListMedia()
	}

	rows, err := db.listBy.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*Media
	for rows.Next() {
		media, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("postgreSQL: could not read row: %v", err)
		}

		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

/*---------------------------  Create/Add  ---------------------------*/

// CreateTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save media, assigning it a new ID.
func (db *pgsqlDB) AddMedia(m *Media) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, m.Title, m.Description,
		m.MediaType, m.Industry, m.ReleaseDate, m.ActorID,
		m.CharacterID, m.DirectorID, m.ImageURL, m.Bechdel,
		m.WikiURL, m.IMDBURL, m.RottenTomURL, m.CreatedByID,
		m.CreatedBy, m.CreatedDate)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("postgreSQL: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}


/*---------------------------  Update  ---------------------------*/

// Updatemedia updates the entry for a given media.
func (db *pgsqlDB) UpdateMedia(m *Media) error {
	if m.ID == 0 {
		return errors.New("postgreSQL: media with unassigned ID passed into update")
	}

	_, err := execAffectingOneRow(db.update, m.Title, m.Description,
		m.MediaType, m.Industry, m.ReleaseDate, m.ActorID,
		m.CharacterID, m.DirectorID, m.ImageURL, m.Bechdel,
		m.WikiURL, m.IMDBURL, m.RottenTomURL, m.CreatedByID,
		m.CreatedBy, m.CreatedDate, m.ID)
	return err
}


/*---------------------------  Delete  ---------------------------*/


// DeleteMedia removes a given book by its ID.
func (db *pgsqlDB) DeleteMedia(id int64) error {
	if id == 0 {
		return errors.New("postgreSQL: media with unassigned ID passed into deleteMedia")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}