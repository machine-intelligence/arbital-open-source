// newSearchStringHandler.go adds a search string to a page
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// newSearchStringData contains data given to us in the request.
type newSearchStringData struct {
	PageID string
	Text   string
}

var newSearchStringHandler = siteHandler{
	URI:         "/newSearchString/",
	HandlerFunc: newSearchStringHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newSearchStringHandlerFunc handles requests to create/update a like.
func newSearchStringHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	c := params.C
	returnData := core.NewHandlerData(u)

	var data newSearchStringData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageID) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}
	if len(data.Text) <= 0 {
		return pages.Fail("Invalid text", nil).Status(http.StatusBadRequest)
	}

	var newId int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Add the new search string
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageID
		hashmap["text"] = data.Text
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("searchStrings", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert into DB", err)
		}
		newId, err = resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get inserted id", err)
		}

		// Update change logs
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.SearchStringChangeChangeLog
		hashmap["newSettingsValue"] = data.Text
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		resp, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to changeLogs", err)
		}
		changeLogId, err := resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get changeLog id", err)
		}

		// Insert updates
		var task tasks.NewUpdateTask
		task.UserId = u.ID
		task.GoToPageId = data.PageID
		task.SubscribedToId = data.PageID
		task.UpdateType = core.ChangeLogUpdateType
		task.ChangeLogId = changeLogId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task: %v", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Update Elastic
	var task tasks.UpdateElasticPageTask
	task.PageID = data.PageID
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	returnData.ResultMap["searchStringId"] = fmt.Sprintf("%d", newId)
	return pages.Success(returnData)
}
