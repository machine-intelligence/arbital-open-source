// updateComment.go can change data for a comment.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	header, err := updateCommentProcessor(c, r)
	if err != nil {
		c.Inc(strings.Trim(r.URL.Path, "/") + "Fail")
		c.Errorf("%v", err)
		w.WriteHeader(header)
		fmt.Fprintf(w, "%v", err)
	}
}

func updateCommentProcessor(c sessions.Context, r *http.Request) (int, error) {
	decoder := json.NewDecoder(r.Body)
	var task updateCommentData
	err := decoder.Decode(&task)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Couldn't decode json: %v", err)
	}
	if task.Text == "" || task.Id <= 0 {
		return http.StatusBadRequest, fmt.Errorf("Invalid parameters")
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = task.Id
	hashmap["text"] = task.Text
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("comments", hashmap, "text", "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Couldn't update comment: %v", err)
	}
	return 0, nil
}
