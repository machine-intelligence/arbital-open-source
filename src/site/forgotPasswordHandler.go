// forgotPasswordHandler.go handles requests when the user says they forgot their password

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/stormpath"
)

// forgotPasswordHandlerData is the data received from the request.
type forgotPasswordHandlerData struct {
	Email string
}

var forgotPasswordHandler = siteHandler{
	URI:         "/json/forgotPassword/",
	HandlerFunc: forgotPasswordHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}

func forgotPasswordHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data forgotPasswordHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if len(data.Email) <= 0 {
		return pages.Fail("Email is not set", nil).Status(http.StatusBadRequest)
	}

	err = stormpath.ForgotPassword(params.C, data.Email)
	if err != nil {
		return pages.Fail("Invalid email or password", err)
	}

	return pages.Success(nil)
}
