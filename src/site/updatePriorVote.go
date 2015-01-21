// updatePriorVote.go can change data for a prior vote.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// updatePriorVoteTask is the object that's put into the daemon queue.
type updatePriorVoteTask struct {
	Id        int64   `json:",string"`
	SupportId int64   `json:",string"`
	Value     float32 `json:",string"`
}

// updatePriorVoteHandler handles requests to create/update a prior vote.
func updatePriorVoteHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updatePriorVoteTask
	err := decoder.Decode(&task)
	if err != nil {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if (task.Id > 0) == (task.SupportId > 0) {
		c.Errorf("Either pass Id to update existing vote OR SupportId to create a new vote")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < 0 || task.Value > 100 {
		c.Errorf("Value has to be between 0 and 100 inclusive")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	updateArgs := make([]string, 0)
	if task.Id > 0 {
		// Updating
		hashmap["id"] = task.Id
		updateArgs = append(updateArgs, "lastChanged", "value")
	} else {
		// Inserting
		var u *user.User
		u, err = user.LoadUser(w, r)
		if err != nil {
			c.Inc("update_prior_vote_fail")
			c.Errorf("Couldn't load user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hashmap["userId"] = u.Id
		hashmap["supportId"] = task.SupportId
	}
	hashmap["lastChanged"] = database.Now()
	hashmap["value"] = task.Value
	sql := database.GetInsertSql("priorVotes", hashmap, updateArgs...)
	if err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_prior_vote_fail")
		c.Errorf("Couldn't update prior vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
