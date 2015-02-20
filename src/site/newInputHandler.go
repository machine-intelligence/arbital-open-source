// newInputHandler.go creates a new input
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// newInputData contains parameters passed in to create a new page
type newInputData struct {
	ParentPageId int64 `json:",string"`
	Url          string
}

// newInputHandler handles requests to create a new page.
func newInputHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newInputData
	err := decoder.Decode(&data)
	if err != nil || len(data.Url) <= 0 || data.ParentPageId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_input_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Parse out pageId and privacyKey from the given url
	exp := regexp.MustCompile("^.*/pages/([[:digit:]]+)/?([[:digit:]]*).*?$")
	results := exp.FindStringSubmatch(data.Url)
	if len(results) <= 1 {
		c.Errorf("Couldn't parse url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pageId := results[1]
	privacyKey := ""
	if len(results) >= 3 {
		privacyKey = results[2]
	}

	// Don't allow to link a page to itself.
	if pageId == fmt.Sprintf("%d", data.ParentPageId) {
		c.Errorf("Trying to link page to itself")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check to see if the linked page is private
	var actualPrivacyKey sql.NullInt64
	found := false
	query := fmt.Sprintf(`
		SELECT privacyKey
		FROM pages
		WHERE id=%s`, pageId)
	found, err = database.QueryRowSql(c, query, &actualPrivacyKey)
	if !found || err != nil {
		c.Errorf("Couldn't load privacyKey: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if actualPrivacyKey.Valid && fmt.Sprintf("%d", actualPrivacyKey.Int64) != privacyKey {
		c.Errorf("The given page is private, but the privacy key is incorrect")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create new input.
	hashmap := make(map[string]interface{})
	hashmap["parentId"] = data.ParentPageId
	hashmap["childId"] = pageId
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	query = database.GetInsertSql("inputs", hashmap)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_input_fail")
		c.Errorf("Couldn't new input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add updates to users who are subscribed to this page.
	var task tasks.NewUpdateTask
	task.UserId = u.Id
	task.PageId = data.ParentPageId
	task.UpdateType = "newInput"
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
