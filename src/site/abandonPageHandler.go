// abandonPageHandler.go handles requests for abandoning a page. This means marking
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

// abandonPageData is the data received from the request.
type abandonPageData struct {
	PageId int64 `json:",string"`
}

// abandonPageHandler handles requests for deleting a page.
func abandonPageHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data abandonPageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		c.Inc("abandon_page_fail")
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
	page, err = loadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		c.Inc("delete_page_fail")
		c.Errorf("Couldn't load page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if page == nil {
		c.Errorf("Couldn't find page: %v", data.PageId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Check that we have the lock
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		c.Inc("delete_page_fail")
		c.Errorf("Don't have the lock")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Get currentEdit number
	var currentEdit int64
	row := db.NewStatement(`
		SELECT ifnull(max(edit), -1)
		FROM pages
		WHERE isCurrentEdit AND pageId=?
		`).QueryRow(data.PageId)
	if _, err = row.Scan(&currentEdit); err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't abandon a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete the edit
	statement := db.NewStatement(`
		UPDATE pages
		SET deletedBy=?
		WHERE pageId=? AND creatorId=? AND isAutosave`)
	if _, err = statement.Exec(u.Id, data.PageId, u.Id); err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't abandon a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["lockedUntil"] = database.Now()
	statement = db.NewInsertStatement("pageInfos", hashmap, "lockedUntil")
	if _, err = statement.Exec(); err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't change lock: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
