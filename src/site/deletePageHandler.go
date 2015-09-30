// deletePageHandler.go handles requests for deleting a page.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// deletePageData is the data received from the request.
type deletePageData struct {
	PageId     int64 `json:",string"`
	UndoDelete bool
}

// deletePageHandler handles requests for deleting a page.
func deletePageHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data deletePageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		c.Inc("delete_page_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("delete_page_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		c.Inc("delete_page_fail")
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
	}
	// Check that we have the lock.
	if page.LockedUntil > database.Now() && page.LockedBy != u.Id {
		c.Inc("delete_page_fail")
		c.Errorf("Don't have the lock")
		w.WriteHeader(http.StatusForbidden)
		return
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
		c.Inc("delete_page_fail")
		c.Errorf("Couldn't delete a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete it from the elastic index
	err = elastic.DeletePageFromIndex(c, data.PageId)
	if err != nil {
		c.Errorf("failed to update index: %v", err)
	}
}
