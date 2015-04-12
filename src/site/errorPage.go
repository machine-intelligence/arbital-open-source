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
	Error string
}

// errorPage serves the "error" page.
var errorPage = pages.Add(
	"/error/",
	errorRender,
	append(baseTmpls, "tmpl/errorPage.tmpl")...)

// everyThing serves the "error" page.
var page404 = pages.Add(
	"/{catchall:.*}", // this is not used
	renderer404,
	append(baseTmpls, "tmpl/errorPage.tmpl")...)

// errorRenderer renders the error page.
func errorRender(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	c.Inc("error_page_served_success")
	return pages.StatusOK(errorData{r.URL.Query().Get("error_msg")})
}

// renderer404 renders the error page.
func renderer404(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	c.Inc("404_page_served_success")
	return pages.CustomCodeWith(errorData{"Page not found :("}, http.StatusNotFound)
}

// showError redirects to the error page, showing a standard message
// and logging specified error.
func showError(w http.ResponseWriter, r *http.Request, err error) *pages.Result {
	next := pages.Values{
		"error_msg": fmt.Sprintf("Error: %v", err),
	}.AddTo(errorPage.URI)
	return pages.RedirectWith(next)
}
