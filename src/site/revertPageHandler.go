// revertPageHandler.go handles requests for reverting a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// revertPageData is the data received from the request.
type revertPageData struct {
	// Page to revert
	PageId int64 `json:",string"`
	// Edit to revert to
	EditNum int
}

// revertPageHandler handles requests for deleting a page.
func revertPageHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data revertPageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Need to be logged in", nil)
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, &loadEditOptions{loadSpecificEdit: data.EditNum})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	} else if page == nil {
		return pages.HandlerErrorFail("Couldn't find page", nil)
	}

	// TODO: check that this user can perform this kind of revertion

	// Check that we have the lock
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		return pages.HandlerErrorFail("Don't have the lock", err)
	}

	// Delete the edit
	statement := db.NewStatement(`
		UPDATE pages
		SET isCurrentEdit=(edit=?)
		WHERE pageId=?`)
	if _, err = statement.Exec(data.EditNum, data.PageId); err != nil {
		return pages.HandlerErrorFail("Couldn't update pages", err)
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["currentEdit"] = data.EditNum
	statement = db.NewInsertStatement("pageInfos", hashmap, "currentEdit")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't change lock", err)
	}
	return pages.StatusOK(nil)
}
