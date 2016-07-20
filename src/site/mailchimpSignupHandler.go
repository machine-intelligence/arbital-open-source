// mailchimpSignupPage.go serves the mailchimpSignup page.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/mailchimp"
	"zanaduu3/src/pages"
)

// mailchimpSignupHandlerData is the data received from the request.
type mailchimpSignupHandlerData struct {
	Email     string
	Interests map[string]bool
}

var mailchimpSignupHandler = siteHandler{
	URI:         "/mailchimpSignup/",
	HandlerFunc: mailchimpSignupHandlerFunc,
	Options:     pages.PageOptions{},
}

func mailchimpSignupHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data mailchimpSignupHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if len(data.Email) <= 0 {
		return pages.Fail("Email have to be specified", nil).Status(http.StatusBadRequest)
	}

	account := &mailchimp.Account{
		Email:     data.Email,
		Interests: data.Interests,
	}

	// Execute request
	err = mailchimp.SubscribeUser(params.C, account)
	if err != nil {
		return pages.Fail("Couldn't subscribe user", err)
	}
	return pages.Success(nil)
}
