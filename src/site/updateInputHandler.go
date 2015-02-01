// updateInput.go can change data for a input.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateInputTask is the object that's put into the daemon queue.
type updateInputTask struct {
	Id   int64 `json:",string"`
	Text string
	Url  string
}

// updateInputHandler renders the input page.
func updateInputHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateInputTask
	err := decoder.Decode(&task)
	if err != nil || task.Id <= 0 || task.Text == "" {
		c.Inc("update_input_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = task.Id
	hashmap["text"] = task.Text
	hashmap["url"] = task.Url
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("inputs", hashmap, "text", "url", "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_input_fail")
		c.Errorf("Couldn't update input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
