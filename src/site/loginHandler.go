// loginPage.go serves the login page.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/okta"
	"zanaduu3/src/pages"
)

// loginHandlerData is the data received from the request.
type loginHandlerData struct {
	Email    string
	Password string
}

var loginHandler = siteHandler{
	URI:         "/login/",
	HandlerFunc: loginHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}
var logoutHandler = siteHandler{
	URI:         "/logout/",
	HandlerFunc: logoutHandlerFunc,
	Options:     pages.PageOptions{},
}

func loginHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data loginHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	return loginHandlerInternalFunc(params, &data)
}

func loginHandlerInternalFunc(params *pages.HandlerParams, data *loginHandlerData) *pages.Result {
	if len(data.Email) <= 0 {
		return pages.Fail("Email has to be specified", nil).Status(http.StatusBadRequest)
	}
	if len(data.Password) <= 0 {
		return pages.Fail("Password has to be specified", nil).Status(http.StatusBadRequest)
	}

	err := okta.AuthenticateUser(params.C, data.Email, data.Password)
	if err != nil {
		return pages.Fail("Invalid email or password", err)
	}
	return setUserInternalFunc(params, data)
}

// Helper function for logging in the user
func setUserInternalFunc(params *pages.HandlerParams, data *loginHandlerData) *pages.Result {
	// Set the cookie
	_, _, err := core.SaveCookie(params.W, params.R, data.Email)
	if err != nil {
		return pages.Fail("Couldn't save a cookie", err)
	}

	// Load the user object from the cookie
	params.U, err = core.LoadCurrentUser(params.W, params.R, params.DB)
	if err != nil {
		return pages.Fail("Couldn't load user", err)
	}

	return pages.Success(core.NewHandlerData(params.U))
}

func logoutHandlerFunc(params *pages.HandlerParams) *pages.Result {
	return pages.Success(nil)
}
