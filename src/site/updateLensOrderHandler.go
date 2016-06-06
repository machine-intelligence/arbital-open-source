// updateLensOrderHandler.go handles reordering of lenses
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

// updateLensOrderData contains the data we get in the request
type updateLensOrderData struct {
	// Id of the page the lenses are for
	PageId string
	// Lens id -> order index map
	OrderMap map[string]int
}

var updateLensOrderHandler = siteHandler{
	URI:         "/updateLensOrder/",
	HandlerFunc: updateLensOrderHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateLensOrderHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateLensOrderData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Page id isn't specified", err).Status(http.StatusBadRequest)
	}
	if len(data.OrderMap) <= 0 {
		return pages.Success(nil)
	}

	// Check permissions
	pageIds := []string{data.PageId}
	for pageId, _ := range data.OrderMap {
		pageIds = append(pageIds, pageId)
	}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err).Status(http.StatusForbidden)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Set up the lens indices
	lensIndexValues := make([]interface{}, 0)
	for pageId, index := range data.OrderMap {
		lensIndexValues = append(lensIndexValues, pageId, index)
	}

	// Begin the transaction.
	var changeLogId int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the lens indices.
		statement := db.NewStatement(`
			INSERT INTO pageInfos (pageId, lensIndex)
			VALUES ` + database.ArgsPlaceholder(len(lensIndexValues), 2) + `
			ON DUPLICATE KEY UPDATE lensIndex=VALUES(lensIndex)`).WithTx(tx)
		if _, err = statement.Exec(lensIndexValues...); err != nil {
			return sessions.NewError("Couldn't update a lens index", err)
		}

		// Create changelogs entry
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.LensOrderChangedChangeLog
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert changeLog", err)
		}
		changeLogId, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get new changeLog id", err)
		}

		// Generate updates for users who are subscribed to the primary page
		var task tasks.NewUpdateTask
		task.UpdateType = core.ChangeLogUpdateType
		task.UserId = u.Id
		task.ChangeLogId = changeLogId
		task.GroupByPageId = data.PageId
		task.SubscribedToId = data.PageId
		task.GoToPageId = data.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
