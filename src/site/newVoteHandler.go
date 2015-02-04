// newVote.go adds a new for for a claim.
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
	redoWindow = 60 // number of seconds during which a user can redo their vote
)

// newVoteTask is the object that's put into the daemon queue.
type newVoteTask struct {
	ClaimId int64 `json:",string"`
	Value   int
}

// newVoteHandler handles requests to create/update a prior vote.
func newVoteHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task newVoteTask
	err := decoder.Decode(&task)
	if err != nil || task.ClaimId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < -1 || task.Value > 1 {
		c.Errorf("Value has to be -1, 0, or 1")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_vote_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check to see if we have a recent vote by this user for this claim.
	var id int64
	var found bool
	query := fmt.Sprintf(`
		SELECT id
		FROM votes
		WHERE TIME_TO_SEC(TIMEDIFF('%s', createdAt)) < %d`, database.Now(), redoWindow)
	found, err = database.QueryRowSql(c, query, &id)
	if err != nil {
		c.Inc("new_vote_fail")
		c.Errorf("Couldn't check for a recent vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if found {
		hashmap := make(map[string]interface{})
		hashmap["id"] = id
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

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["claimId"] = task.ClaimId
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
