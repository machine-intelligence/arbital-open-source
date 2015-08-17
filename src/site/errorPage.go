// error.go: static error page.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

type errorData struct {
	commonPageData
	Error string
}

// errorPage serves the "error" page.
var errorPage = newPage(
	"/error/",
	errorRender,
	append(baseTmpls,
		"tmpl/errorPage.tmpl", "tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// everyThing serves the "error" page.
var page404 = newPage(
	"/{catchall:.*}", // this is not used
	renderer404,
	append(baseTmpls, "tmpl/errorPage.tmpl"))

// errorRenderer renders the error page.
func errorRender(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	c.Inc("error_page_served_success")
	var data errorData
	data.Error = r.URL.Query().Get("error_msg")
	data.User = u
	return pages.StatusOK(data)
}

// renderer404 renders the error page.
func renderer404(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	c.Inc("404_page_served_success")
	return showError(w, r, fmt.Errorf("Page not found :("))
}

// showError redirects to the error page, showing a standard message
// and logging specified error.
func showError(w http.ResponseWriter, r *http.Request, err error) *pages.Result {
	next := pages.Values{
		"error_msg": fmt.Sprintf("Error: %v", err),
	}.AddTo(errorPage.URI)
	return pages.RedirectWith(next)
}
