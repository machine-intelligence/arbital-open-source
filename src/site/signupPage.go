// signupPage.go serves the signup page.
package site

import (
	"zanaduu3/src/pages"
)

// signupData stores the data that we pass to the signup.tmpl to render the page
type signupData struct {
	commonPageData
	ContinueUrl string
}

// signupPage serves the signup page.
var signupPage = newPageWithOptions(
	"/signup/",
	signupRenderer,
	append(baseTmpls, "tmpl/angular.tmpl.js",
		"tmpl/signupPage.tmpl"),
	pages.PageOptions{})

// signupRenderer renders the signup page.
func signupRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U

	if u.Id <= 0 {
		return pages.RedirectWith(u.LoginLink)
	}

	var data signupData
	data.User = u
	data.ContinueUrl = params.R.URL.Query().Get("continueUrl")
	return pages.StatusOK(data)
}
