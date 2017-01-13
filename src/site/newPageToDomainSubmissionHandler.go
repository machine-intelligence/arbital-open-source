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
	PageID   string `json:"pageId"`
	DomainID string `json:"domainId"`
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
	u := params.U
	returnData := core.NewHandlerData(u)

	var data newPageToDomainSubmissionData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return CreatePageToDomainSubmission(tx, u, &data)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	returnData.ResultMap["submission"], err = core.LoadPageToDomainSubmission(db, data.PageID, data.DomainID)
	if err != nil {
		return pages.Fail("Couldn't load submission", err)
	}
	return pages.Success(returnData)
}

func CreatePageToDomainSubmission(tx *database.Tx, u *core.CurrentUser, data *newPageToDomainSubmissionData) sessions.Error {
	// Create new submission
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageID
	hashmap["domainId"] = data.DomainID
	hashmap["submitterId"] = u.ID
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("pageToDomainSubmissions", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't add submission", err)
	}

	// Notify all domain owners about this new submission
	var task tasks.DomainWideNewUpdateTask
	task.UserID = u.ID
	task.UpdateType = core.PageToDomainSubmissionUpdateType
	task.DomainID = data.DomainID
	task.GoToPageID = data.PageID
	if err := tasks.Enqueue(tx.DB.C, &task, nil); err != nil {
		return sessions.NewError("Couldn't enqueue a task", err)
	}
	return nil
}
