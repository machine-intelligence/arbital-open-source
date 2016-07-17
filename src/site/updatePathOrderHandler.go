// updatePathOrderHandler.go handles reordering of pages in a path
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

// updatePathOrderData contains the data we get in the request
type updatePathOrderData struct {
	// Id of the page the path pages are for
	GuideId string
	// Map of ids -> path index
	PageOrder map[string]int
}

var updatePathOrderHandler = siteHandler{
	URI:         "/json/updatePathOrder/",
	HandlerFunc: updatePathOrderHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updatePathOrderHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updatePathOrderData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.GuideId) {
		return pages.Fail("Guide id isn't valid", nil).Status(http.StatusBadRequest)
	}

	// Load all the path pages
	pathPages := make([]*core.PathPage, 0)
	queryPart := database.NewQuery(`WHERE pathp.guideId=?`, data.GuideId)
	err = core.LoadPathPages(db, queryPart, nil, func(db *database.DB, pathPage *core.PathPage) error {
		pathPages = append(pathPages, pathPage)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the path pages: %v", err)
	} else if len(pathPages) <= 0 {
		return pages.Fail("No path pages found for this guide", nil).Status(http.StatusBadRequest)
	}

	// Check permissions
	pageIds := []string{data.GuideId}
	permissionError, err := core.VerifyEditPermissionsForList(db, pageIds, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Set up the path indices
	hashmaps := make(database.InsertMaps, 0)
	for _, pathPage := range pathPages {
		hashmap := make(database.InsertMap)
		hashmap["id"] = pathPage.ID
		hashmap["pathIndex"] = data.PageOrder[fmt.Sprintf("%d", pathPage.ID)]
		hashmap["updatedBy"] = u.ID
		hashmap["updatedAt"] = database.Now()
		hashmaps = append(hashmaps, hashmap)
	}

	// Begin the transaction.
	var changeLogId int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the pathPages
		statement := db.NewMultipleInsertStatement("pathPages", hashmaps, "pathIndex", "updatedBy", "updatedAt").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pathPages", err)
		}

		// Create changelogs entry
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.GuideId
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.PathOrderChangedChangeLog
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
		task.UserId = u.ID
		task.ChangeLogId = changeLogId
		task.SubscribedToId = data.GuideId
		task.GoToPageId = data.GuideId
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
