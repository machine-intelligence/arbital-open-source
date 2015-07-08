// abandonPageHandler.go handles requests for abandoning a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
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

	// Get currentEdit number.
	var currentEdit int64
	query := fmt.Sprintf(`SELECT ifnull(max(edit), -1) FROM pages WHERE isCurrentEdit AND pageId=%d`, data.PageId)
	if _, err = database.QueryRowSql(c, query, &currentEdit); err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't abandon a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	query = fmt.Sprintf(`
		UPDATE pages
		SET deletedBy=%d
		WHERE pageId=%d AND edit > %d AND creatorId=%d AND (isSnapshot || isAutosave)`,
		u.Id, data.PageId, currentEdit, u.Id)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("abandon_page_fail")
		c.Errorf("Couldn't abandon a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
