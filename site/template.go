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
"fmt"
"html/template"
"io/ioutil"
"net/http"
"path/filepath"
	"time"
)

/*TODO add back bookshelf and update for new data cols avail*/

// parseTemplate applies a given file to the body of the base template.
func parseTemplate(filename string) *appTemplate {
	if debugProject { fmt.Sprintf("PARSING") }

	pagePath :="content/startbootstrap/"

	tmpl := template.Must(template.ParseFiles(pagePath + "base.html"))

	// Put the named file into a template called "body"
	path := filepath.Join(pagePath, filename)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("could not read template: %v", err))
	}
	template.Must(tmpl.New("body").Parse(string(b)))

	if false {
		pagePath ="content/bookshelf/"
		tmpl = template.Must(template.ParseFiles(pagePath + "base.html"))
		return &appTemplate{tmpl.Lookup("index.html")}
	}
	return &appTemplate{tmpl.Lookup("base.html")}
}

// appTemplate is a user login-aware wrapper for a html/template.
type appTemplate struct {
	t *template.Template
}

// Execute writes the template using the provided data, adding login and user
// information to the base template.
func (tmpl *appTemplate) Execute(w http.ResponseWriter, r *http.Request, data interface{}) error {
	d := struct {
		PageTitle   string
		Date 		string
		Data        interface{}
	}{
		PageTitle:	"Flip the Script",
		Date:		time.Now().Format("02-01-2006"),
		Data:		data,
	}
	if err := tmpl.t.Execute(w, d); err != nil {
		return appErrorf(err, "could not write template: %v", err)
	}
	return nil
}