// oops.go: static error page.
package site

import (
	"net/http"

	"zanaduu3/src/sessions"

	"github.com/hkjn/pages"
)

// oopsPage serves the "oops" page.
var oopsPage = pages.Add(
	"/oops",
	func(w http.ResponseWriter, r *http.Request) *pages.Result {
		c := sessions.NewContext(r)
		c.Inc("oops_page_served_success")
		return pages.StatusOK(
			struct {
				ErrorMsg string
			}{r.URL.Query().Get("error_msg")})
	},
	append(baseTmpls, "tmpl/base.tmpl", "tmpl/oops.tmpl")...)

// showOops redirects to the "oops" page, showing a standard message
// and logging specified error.
func showOops(w http.ResponseWriter, r *http.Request, err error) {
	c := sessions.NewContext(r)
	c.Errorf("redirecting to %s: %v\n", oopsPage.URI, err)
	next := pages.Values{
		"error_msg": "There was a temporary problem. Please try again later.",
	}.AddTo(oopsPage.URI)
	http.Redirect(w, r, next, http.StatusSeeOther)
}
