// updateCommentVoteHandler.go adds or updates a comment vote.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// updateCommentVoteData is the object that's put into the daemon queue.
type updateCommentVoteData struct {
	CommentId int64 `json:",string"`
	Value     int
}

// updateCommentVoteHandler handles requests to create/update a prior vote.
func updateCommentVoteHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateCommentVoteData
	err := decoder.Decode(&task)
	if err != nil || task.CommentId <= 0 {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if task.Value < 0 || task.Value > 1 {
		c.Errorf("Value has to be 0 or 1")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_comment_vote_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Update comment vote
	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["commentId"] = task.CommentId
	hashmap["value"] = task.Value
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	query := database.GetInsertSql("commentVotes", hashmap, "value", "updatedAt")
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_comment_vote_fail")
		c.Errorf("Couldn't add a comment vote: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
