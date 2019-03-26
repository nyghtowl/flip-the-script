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

// Media holds metadata about a Media.
type Media struct {
	ID            int64
	Title         string
	Description   string
	MediaType 	  string
	Industry	  string
	ReleaseDate	  string

	ActorID		  int64
	CharacterID	  int64
	DirectorID	  int64

	ImageURL	  string
	Bechdel		  bool
	WikiURL		  string
	IMDBURL		  string
	RottenTomURL  string

	CreatedByID	  int64
	CreatedBy     string
	CreatedDate	  string

	PageSubTitle  string

}

// CreatedByDisplayName returns a string appropriate for displaying the name of
// the user who created this media object.
func (m *Media) CreatedByDisplayName() string {
	if m.CreatedBy == "" {
		return "Anonymous"
	}
	return m.CreatedBy
}

// SetCreatorAnonymous sets the CreatedByID field to the "anonymous" ID.
func (m *Media) SetCreatorAnonymous() {
	m.CreatedBy = "anonymous"
	m.CreatedByID = 0000
}

// MediaDatabase provides thread-safe access to a database of media.
type MediaDatabase interface {
	// ListMedia returns a list of Medias, ordered by title.
	ListMedia() ([]*Media, error)

	// ListMediaCreatedBy returns a list of media, ordered by title, filtered by
	// the user who created the Media entry.
	ListMediaCreatedBy(userID int64) ([]*Media, error)

	// GetMedia retrieves a Media by its ID.
	GetMedia(id int64) (*Media, error)

	// AddMedia saves a given media, assigning it a new ID.
	AddMedia(m *Media) (id int64, err error)

	// DeleteMedia removes a given media by its ID.
	DeleteMedia(id int64) error

	// UpdateMedia updates the entry for a given media.
	UpdateMedia(m *Media) error

	// Close closes the database, freeing up any available resources.
	Close()
}