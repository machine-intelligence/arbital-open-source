// priorVote.go can change data for a prior vote.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// priorVoteTask is the object that's put into the daemon queue.
type priorVoteTask struct {
	QuestionId int64   `json:",string"`
	Value      float32 `json:",string"`
}

// priorVoteHandler handles requests to create/update a prior vote.
func priorVoteHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task priorVoteTask
	err := decoder.Decode(&task)
	if err != nil || task.QuestionId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < 0 || task.Value > 100 {
		c.Errorf("Value has to be between 0 and 100 inclusive")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("update_prior_vote_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["questionId"] = task.QuestionId
	hashmap["createdAt"] = database.Now()
	hashmap["value"] = task.Value
	sql := database.GetInsertSql("priorVotes", hashmap)
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_prior_vote_fail")
		c.Errorf("Couldn't add a prior vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
