// revertPageHandler.go handles requests for reverting a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// revertPageData is the data received from the request.
type revertPageData struct {
	// Page to revert
	PageId int64 `json:",string"`
	// Edit to revert to
	EditNum int
}

// revertPageHandler handles requests for deleting a page.
func revertPageHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data revertPageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, &loadEditOptions{loadSpecificEdit: data.EditNum})
	if err != nil {
		c.Errorf("Couldn't load page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if page == nil {
		c.Errorf("Couldn't find page: %v", data.PageId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: check that this user can perform this kind of revertion

	// Check that we have the lock
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		c.Errorf("Don't have the lock")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Delete the edit
	statement := db.NewStatement(`
		UPDATE pages
		SET isCurrentEdit=(edit=?)
		WHERE pageId=?`)
	if _, err = statement.Exec(data.EditNum, data.PageId); err != nil {
		c.Errorf("Couldn't update pages: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["currentEdit"] = data.EditNum
	statement = db.NewInsertStatement("pageInfos", hashmap, "currentEdit")
	if _, err = statement.Exec(); err != nil {
		c.Errorf("Couldn't change lock: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
