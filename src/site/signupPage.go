// signupPage.go serves the signup page.
package site

import (
	"zanaduu3/src/pages"
)

// signupPage serves the signup page.
var signupPage = newPage(signupRenderer, dynamicTmpls)

// signupRenderer renders the signup page.
func signupRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U

	if u.Id <= 0 {
		return pages.RedirectWith(u.LoginLink)
	}

	return pages.StatusOK(nil)
}
