// newVoteHadler.go adds a new vote for for a page.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newVoteData contains data given to us in the request.
type newVoteData struct {
	PageId int64 `json:",string"`
	Value  float32
}

// newVoteHandler handles requests to create/update a prior vote.
func newVoteHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task newVoteData
	err := decoder.Decode(&task)
	if err != nil || task.PageId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < 0 || task.Value >= 100 {
		c.Errorf("Value has to be [0, 100)")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_vote_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Get the last vote.
	var oldVoteId int64
	var oldVoteValue float32
	var oldVoteExists bool
	var oldVoteAge int64
	query := fmt.Sprintf(`
		SELECT id,value,TIME_TO_SEC(TIMEDIFF('%s',createdAt)) AS age
		FROM votes
		WHERE userId=%d AND pageId=%d
		ORDER BY id DESC
		LIMIT 1`,
		database.Now(), u.Id, task.PageId)
	oldVoteExists, err = database.QueryRowSql(c, query, &oldVoteId, &oldVoteValue, &oldVoteAge)
	if err != nil {
		c.Inc("new_vote_fail")
		c.Errorf("Couldn't check for a recent vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If previous vote is exactly the same, don't do anything.
	if oldVoteExists && oldVoteValue == task.Value {
		return
	}

	// Check to see if we have a recent vote by this user for this page.
	if oldVoteExists && oldVoteAge <= redoWindow {
		hashmap := make(map[string]interface{})
		hashmap["id"] = oldVoteId
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		query = database.GetInsertSql("votes", hashmap, "value", "createdAt")
		if _, err = database.ExecuteSql(c, query); err != nil {
			c.Inc("new_vote_fail")
			c.Errorf("Couldn't update a vote: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Insert new vote.
	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["pageId"] = task.PageId
	hashmap["value"] = task.Value
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("votes", hashmap)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_vote_fail")
		c.Errorf("Couldn't add a vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
