// index.go serves the index page.
package site

import (
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	ErrorMsg string
}

// indexPage serves the index page.
var indexPage = pages.Add(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/base.tmpl",
		"tmpl/main.tmpl")...)

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data indexTmplData
	c := sessions.NewContext(r)

	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}
