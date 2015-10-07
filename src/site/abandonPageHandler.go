// abandonPageHandler.go handles requests for abandoning a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// abandonPageData is the data received from the request.
type abandonPageData struct {
	PageId int64 `json:",string"`
}

// abandonPageHandler handles requests for deleting a page.
func abandonPageHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	decoder := json.NewDecoder(params.R.Body)
	var data abandonPageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	} else if page == nil {
		return pages.HandlerBadRequestFail("Couldn't find page", nil)
	}
	// Check that we have the lock
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		return pages.HandlerForbiddenFail("Don't have the lock", nil)
	}

	// Get currentEdit number
	var currentEdit int64
	row := db.NewStatement(`
		SELECT ifnull(max(edit), -1)
		FROM pages
		WHERE isCurrentEdit AND pageId=?
		`).QueryRow(data.PageId)
	if _, err = row.Scan(&currentEdit); err != nil {
		return pages.HandlerErrorFail("Couldn't abandon a page", err)
	}

	// Delete the edit
	statement := db.NewStatement(`
		UPDATE pages
		SET deletedBy=?
		WHERE pageId=? AND creatorId=? AND isAutosave`)
	if _, err = statement.Exec(u.Id, data.PageId, u.Id); err != nil {
		return pages.HandlerErrorFail("Couldn't abandon a page", err)
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["lockedUntil"] = database.Now()
	statement = db.NewInsertStatement("pageInfos", hashmap, "lockedUntil")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't change lock", err)
	}
	return pages.StatusOK(nil)
}
