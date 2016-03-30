// mailchimpSignupPage.go serves the mailchimpSignup page.
package site

import (
	"encoding/json"

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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if len(data.Email) <= 0 {
		return pages.HandlerBadRequestFail("Email have to be specified", nil)
	}

	account := &mailchimp.Account{
		Email:     data.Email,
		Interests: data.Interests,
	}

	// Execute request
	err = mailchimp.SubscribeUser(params.C, account)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't subscribe user", err)
	}
	return pages.StatusOK(nil)
}
