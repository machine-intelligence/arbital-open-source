// updateLensOrderHandler.go handles reordering of lenses

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

// updateLensOrderData contains the data we get in the request
type updateLensOrderData struct {
	// Id of the page the lenses are for
	PageID string
	// Map of ids -> lens index
	LensOrder map[string]int
}

var updateLensOrderHandler = siteHandler{
	URI:         "/json/updateLensOrder/",
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
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("PageId isn't valid", nil).Status(http.StatusBadRequest)
	}

	// Load all the lenses
	pageMap := make(map[string]*core.Page)
	lenses := make([]*core.Lens, 0)
	queryPart := database.NewQuery(`WHERE l.pageId=?`, data.PageID)
	err = core.LoadLenses(db, queryPart, nil, func(db *database.DB, lens *core.Lens) error {
		lenses = append(lenses, lens)
		core.AddPageIDToMap(lens.PageID, pageMap)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load the lens: %v", err)
	} else if len(lenses) <= 0 {
		return pages.Fail("No lenses found for this page", nil).Status(http.StatusBadRequest)
	} else if len(pageMap) > 1 {
		return pages.Fail("Changing lenses that belong to different pages", nil).Status(http.StatusBadRequest)
	}

	// Check permissions
	permissionError, err := core.VerifyEditPermissionsForMap(db, pageMap, u)
	if err != nil {
		return pages.Fail("Error verifying permissions", err).Status(http.StatusForbidden)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Set up the lens indices
	hashmaps := make(database.InsertMaps, 0)
	for _, lens := range lenses {
		hashmap := make(database.InsertMap)
		hashmap["id"] = lens.ID
		hashmap["lensIndex"] = data.LensOrder[fmt.Sprintf("%d", lens.ID)]
		hashmap["updatedBy"] = u.ID
		hashmap["updatedAt"] = database.Now()
		hashmaps = append(hashmaps, hashmap)
	}

	// Begin the transaction.
	var changeLogID int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the lenses
		statement := db.NewMultipleInsertStatement("lenses", hashmaps, "lensIndex", "updatedBy", "updatedAt").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update lenses", err)
		}

		// Create changelogs entry
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.LensOrderChangedChangeLog
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert changeLog", err)
		}
		changeLogID, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get new changeLog id", err)
		}

		// Generate updates for users who are subscribed to the primary page
		var task tasks.NewUpdateTask
		task.UpdateType = core.ChangeLogUpdateType
		task.UserID = u.ID
		task.ChangeLogID = changeLogID
		task.SubscribedToID = data.PageID
		task.GoToPageID = data.PageID
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
