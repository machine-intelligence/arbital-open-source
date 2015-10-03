// deletePageHandler.go handles requests for deleting a page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
)

// deletePageData is the data received from the request.
type deletePageData struct {
	PageId     int64 `json:",string"`
	UndoDelete bool
}

// deletePageHandler handles requests for deleting a page.
func deletePageHandler(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data deletePageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	}
	// Check that we have the lock.
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		return pages.HandlerForbiddenFail("Don't have the lock", nil)
	}

	// Perform delete.
	deletedBy := u.Id
	if data.UndoDelete {
		deletedBy = 0
	}
	statement := db.NewStatement(`
		UPDATE pages
		SET deletedBy=?
		WHERE pageId=?`)
	if _, err = statement.Exec(deletedBy, data.PageId); err != nil {
		return pages.HandlerErrorFail("Couldn't delete a page", err)
	}

	// Delete it from the elastic index
	err = elastic.DeletePageFromIndex(c, data.PageId)
	if err != nil {
		return pages.HandlerErrorFail("failed to update index", err)
	}
	return pages.StatusOK(nil)
}
