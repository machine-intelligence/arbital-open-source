// loginPage.go serves the login page.
package site

import (
	"encoding/json"

	"zanaduu3/src/pages"
	"zanaduu3/src/stormpath"
	"zanaduu3/src/user"
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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if len(data.Email) <= 0 || len(data.Password) <= 0 {
		return pages.HandlerBadRequestFail("Email and password have to be specified", nil)
	}

	err = stormpath.AuthenticateUser(params.C, data.Email, data.Password)
	if err != nil {
		return pages.HandlerErrorFail("Invalid email or password", err)
	}

	// Set the cookie
	err = user.SaveEmailCookie(params.W, params.R, data.Email)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't save a cookie", err)
	}

	return pages.StatusOK(nil)
}

func logoutHandlerFunc(params *pages.HandlerParams) *pages.Result {
	return pages.StatusOK(nil)
}
