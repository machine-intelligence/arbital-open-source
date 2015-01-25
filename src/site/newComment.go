// newComment.go can change data for a comment.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newCommentTask is the object that's put into the daemon queue.
type newCommentTask struct {
	Text      string
	InputId   int64 `json:",string"`
	ReplyToId int64 `json:",string"`
}

// newCommentHandler renders the comment page.
func newCommentHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task newCommentTask
	err := decoder.Decode(&task)
	if err != nil || task.Text == "" {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["inputId"] = task.InputId
	hashmap["createdAt"] = database.Now()
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	if task.ReplyToId > 0 {
		hashmap["replyToId"] = task.ReplyToId
	}
	hashmap["text"] = task.Text
	sql := database.GetInsertSql("comments", hashmap)
	if err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't create new comment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
