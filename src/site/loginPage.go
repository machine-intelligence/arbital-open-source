// loginPage.go serves the login page.
package site

import (
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// loginData stores the data that we pass to the login.tmpl to render the page
type loginData struct {
	ContinueUri string
}

// loginPage serves the questions page.
var loginPage = pages.Add(
	"/login/",
	loginRenderer,
	append(baseTmpls,
		"tmpl/login.tmpl", "tmpl/navbar.tmpl")...)

// loginRenderer renders the login page.
func loginRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	c := sessions.NewContext(r)

	q := r.URL.Query()
	data := loginData{ContinueUri: q.Get("continueUri")}
	c.Inc("login_page_served_success")
	return pages.StatusOK(data)
}
