// deletePageHandler.go handles requests for deleting a page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// deletePageData is the data received from the request.
type deletePageData struct {
	PageId int64 `json:",string"`
}

var deletePageHandler = siteHandler{
	URI:         "/deletePage/",
	HandlerFunc: deletePageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// deletePageHandlerFunc handles requests for deleting a page.
func deletePageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	decoder := json.NewDecoder(params.R.Body)
	var data deletePageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.PageId == 0 {
		return pages.HandlerBadRequestFail("PageId isn't set", nil)
	}

	// Load the page
	pageMap := make(map[int64]*core.Page)
	page := core.AddPageIdToMap(data.PageId, pageMap)
	err = core.LoadPages(db, u, pageMap)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	}
	if page == nil {
		// Looks like there is no need to delete this page.
		return pages.StatusOK(nil)
	}
	if page.Type == core.GroupPageType || page.Type == core.DomainPageType {
		if !u.IsAdmin {
			return pages.HandlerForbiddenFail("Have to be an admin to delete a group/domain", nil)
		}
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Delete all pairs
		rows := database.NewQuery(`
			SELECT parentId,childId,type
			FROM pagePairs
			WHERE parentId=? OR childId=?`, data.PageId, data.PageId).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId int64
			var pairType string
			err := rows.Scan(&parentId, &childId, &pairType)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
				return deletePagePair(tx, u.Id, parentId, childId, pairType)
			})
			if errMessage != "" {
				return fmt.Errorf("%s: %v", errMessage, err)
			}
			return nil
		})
		if err != nil {
			return "Couldn't load pairs: %v", err
		}

		// Clear the current edit in pages
		statement := tx.NewTxStatement("UPDATE pages SET isCurrentEdit=false WHERE pageId=? AND isCurrentEdit")
		if _, err = statement.Exec(data.PageId); err != nil {
			return "Couldn't update isCurrentEdit for old edits", err
		}

		// Update pageInfos table
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["currentEdit"] = 0
		statement = tx.NewInsertTxStatement("pageInfos", hashmap, "currentEdit")
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update pageInfos", err
		}

		// Update change log
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.DeletePageChangeLog
		statement = tx.NewInsertTxStatement("changeLogs", hashmap)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update change logs", err
		}

		// Delete it from the elastic index
		err = elastic.DeletePageFromIndex(params.C, data.PageId)
		if err != nil {
			return "failed to update index", err
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(fmt.Sprintf("Transaction failed: %s", errMessage), err)
	}

	// Create a task to propagate the domain change to all children
	var task tasks.PropagateDomainTask
	task.PageId = data.PageId
	task.Deleted = true
	if err := task.IsValid(); err != nil {
		return pages.HandlerErrorFail("Invalid task created: %v", err)
	} else if err := tasks.Enqueue(params.C, task, "propagateDomain"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
