// index.go serves the index page.
package site

import (
	//"html/template"
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
var indexPage = newPage(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/index.tmpl",
		"tmpl/navbar.tmpl"))

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var data indexTmplData
	data.User = u
	c := sessions.NewContext(r)

	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}
