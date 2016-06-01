// newPageToDomainSubmissionHandler.go adds new invites to db and auto-claims / sends invite emails
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// updateSettingsData contains data given to us in the request.
type newPageToDomainSubmissionData struct {
	PageId string `json:"pageId"`
}

var newPageToDomainSubmissionHandler = siteHandler{
	URI:         "/json/newPageToDomainSubmission/",
	HandlerFunc: newPageToDomainSubmissionHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateSettingsHandlerFunc handles submitting the settings from the Settings page
func newPageToDomainSubmissionHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U
	returnData := core.NewHandlerData(u)

	var data newPageToDomainSubmissionData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		// Create new invite
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["domainId"] = "1lw"
		hashmap["submitterId"] = u.Id
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("pageToDomainSubmissions", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't add submission", err)
		}

		// Notify all domain owners about this new submission
		var task tasks.DomainWideNewUpdateTask
		task.UserId = u.Id
		task.UpdateType = core.PageToDomainSubmissionUpdateType
		task.DomainId = "1lw"
		task.GoToPageId = data.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	returnData.ResultMap["submission"], err = core.LoadPageToDomainSubmission(db, data.PageId, "1lw")
	if err != nil {
		return pages.Fail("Couldn't load submission", err)
	}
	return pages.Success(returnData)
}
