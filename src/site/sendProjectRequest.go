// sendProjectRequestHandler.go sends an email to us about a project request a person wants

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine/mail"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// sendProjectRequestData contains data given to us in the request.
type sendProjectRequestData struct {
	Text string
}

var sendProjectRequestHandler = siteHandler{
	URI:         "/json/sendProjectRequest/",
	HandlerFunc: sendProjectRequestHandlerFunc,
	Options:     pages.PageOptions{},
}

// sendProjectRequestHandlerFunc handles requests to create/update a like.
func sendProjectRequestHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	c := params.C

	var data sendProjectRequestData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if data.Text == "" {
		return pages.Fail("Text isn't set", nil).Status(http.StatusBadRequest)
	}

	if sessions.Live {
		// Create mail message
		msg := &mail.Message{
			Sender:  "alexei@arbital.com",
			To:      []string{"trigger@recipe.ifttt.com"},
			Subject: "#slackbot",
			Body:    fmt.Sprintf("(id: %s) made a request:\n%s", u.ID, data.Text),
		}

		err = mail.Send(c, msg)
		if err != nil {
			return pages.Fail("Couldn't send email: %v", err)
		}
	} else {
		// If not live, then do nothing, for now
	}

	return pages.Success(nil)
}
