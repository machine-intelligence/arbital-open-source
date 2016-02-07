// signupPage.go serves the signup page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/user"
)

// signupPage serves the signup page.
var signupPage = newPage(signupRenderer, dynamicTmpls)
var loginPage = newPage(loginRenderer, dynamicTmpls)
var logoutPage = newPage(logoutRenderer, dynamicTmpls)

// signupRenderer renders the signup page.
func signupRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U

	if !core.IsIdValid(u.Id) {
		loginLink, err := user.GetLoginLink(params.C, params.R.FormValue("continueUrl"))
		if err != nil {
			pages.Fail("Couldn't get login link", err)
		}
		return pages.RedirectWith(loginLink)
	}

	return pages.StatusOK(nil)
}

// loginRenderer renders the signup page.
func loginRenderer(params *pages.HandlerParams) *pages.Result {
	loginLink, err := user.GetLoginLink(params.C, params.R.FormValue("continueUrl"))
	if err != nil {
		pages.Fail("Couldn't get login link", err)
	}
	return pages.RedirectWith(loginLink)
}

// logoutRenderer renders the signup page.
func logoutRenderer(params *pages.HandlerParams) *pages.Result {
	logoutLink, err := user.GetLogoutLink(params.C, params.R.FormValue("continueUrl"))
	if err != nil {
		pages.Fail("Couldn't get logout link", err)
	}
	return pages.RedirectWith(logoutLink)
}
