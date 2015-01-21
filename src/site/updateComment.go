// updateComment.go can change data for a comment.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// updateCommentTask is the object that's put into the daemon queue.
type updateCommentTask struct {
	Id        int64 `json:",string"`
	Text      string
	SupportId int64 `json:",string"`
	ReplyToId int64 `json:",string"`
}

// updateCommentHandler renders the comment page.
func updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateCommentTask
	err := decoder.Decode(&task)
	if err != nil || task.Text == "" {
		c.Inc("update_comment_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	updateArgs := make([]string, 0)
	if task.Id > 0 {
		// Updating
		hashmap["id"] = task.Id
	} else {
		// Inserting
		var u *user.User
		u, err = user.LoadUser(w, r)
		if err != nil {
			c.Inc("update_comment_fail")
			c.Errorf("Couldn't load user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hashmap["createdAt"] = database.Now()
		hashmap["creatorId"] = u.Id
		hashmap["creatorName"] = u.FullName()
	}
	if task.SupportId > 0 {
		hashmap["supportId"] = task.SupportId
	}
	if task.ReplyToId > 0 {
		hashmap["replyToId"] = task.ReplyToId
	}
	hashmap["text"] = task.Text
	updateArgs = append(updateArgs, "text")
	sql := database.GetInsertSql("comments", hashmap, updateArgs...)
	if err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_comment_fail")
		c.Errorf("Couldn't update comment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
