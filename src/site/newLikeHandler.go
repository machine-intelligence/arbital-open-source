// newLikeHadler.go adds a new like for for a page.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	redoWindow = 60 // number of seconds during which a user can redo their like
)

// newLikeData contains data given to us in the request.
type newLikeData struct {
	PageId int64 `json:",string"`
	Value  int
}

// newLikeHandler handles requests to create/update a prior like.
func newLikeHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task newLikeData
	err := decoder.Decode(&task)
	if err != nil || task.PageId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < -1 || task.Value > 1 {
		c.Errorf("Value has to be -1, 0, or 1")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_like_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check to see if we have a recent like by this user for this page.
	var id int64
	var found bool
	query := fmt.Sprintf(`
		SELECT id
		FROM likes
		WHERE userId=%d AND pageId=%d AND TIME_TO_SEC(TIMEDIFF('%s', createdAt)) < %d`,
		u.Id, task.PageId, database.Now(), redoWindow)
	found, err = database.QueryRowSql(c, query, &id)
	if err != nil {
		c.Inc("new_like_fail")
		c.Errorf("Couldn't check for a recent like: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if found {
		hashmap := make(map[string]interface{})
		hashmap["id"] = id
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		query = database.GetInsertSql("likes", hashmap, "value", "createdAt")
		if _, err = database.ExecuteSql(c, query); err != nil {
			c.Inc("new_like_fail")
			c.Errorf("Couldn't update a like: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["pageId"] = task.PageId
	hashmap["value"] = task.Value
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("likes", hashmap)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_like_fail")
		c.Errorf("Couldn't add a like: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
