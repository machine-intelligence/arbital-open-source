// updateComment.go can change data for a comment.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateCommentData is the object that's put into the daemon queue.
type updateCommentData struct {
	Id   int64 `json:",string"`
	Text string
}

// updateCommentHandler renders the comment page.
func updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateCommentData
	err := decoder.Decode(&task)
	if err != nil || task.Text == "" || task.Id <= 0 {
		c.Inc("update_comment_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = task.Id
	hashmap["text"] = task.Text
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("comments", hashmap, "text", "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_comment_fail")
		c.Errorf("Couldn't update comment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
