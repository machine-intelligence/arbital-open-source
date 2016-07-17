// feedbackHandler.go adds a new vote for for a page.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// feedbackData contains data given to us in the request.
type feedbackData struct {
	Text string
}

var feedbackHandler = siteHandler{
	URI:         "/feedback/",
	HandlerFunc: feedbackHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// feedbackHandlerFunc handles requests to create/update a prior vote.
func feedbackHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	c := params.C

	decoder := json.NewDecoder(params.R.Body)
	var data feedbackData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if data.Text == "" {
		return pages.Fail("No text specified", nil).Status(http.StatusBadRequest)
	}

	var task tasks.SendFeedbackEmailTask
	task.UserId = u.ID
	task.UserEmail = u.Email
	task.Text = data.Text
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
