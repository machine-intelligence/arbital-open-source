// index.go serves the index page.
package site

import (
	"html/template"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	User *user.User
}

// indexPage serves the index page.
var indexPage = pages.Add(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/index.tmpl",
		"tmpl/navbar.tmpl")...)

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data indexTmplData
	c := sessions.NewContext(r)

	// Load user, if possible
	var err error
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"UserId":  func() int64 { return data.User.Id },
		"IsAdmin": func() bool { return data.User.IsAdmin },
	}
	c.Inc("index_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
