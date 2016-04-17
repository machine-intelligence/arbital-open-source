// newSearchStringHandler.go adds a search string to a page
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// newSearchStringData contains data given to us in the request.
type newSearchStringData struct {
	PageId string
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
	returnData := core.NewHandlerData(params.U, false)

	var data newSearchStringData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if len(data.Text) <= 0 {
		return pages.HandlerBadRequestFail("Invalid text", nil)
	}

	var newId int64
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Add the new search string
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.PageId
		hashmap["text"] = data.Text
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("searchStrings", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return "Couldn't insert into DB", err
		}

		newId, err = resp.LastInsertId()
		if err != nil {
			return "Couldn't get inserted id", err
		}

		// Update change logs
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.SearchStringChangeChangeLog
		hashmap["newSettingsValue"] = data.Text
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't add to changeLogs", err
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Update Elastic
	var task tasks.UpdateElasticPageTask
	task.PageId = data.PageId
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	returnData.ResultMap["searchStringId"] = fmt.Sprintf("%d", newId)
	return pages.StatusOK(returnData)
}
