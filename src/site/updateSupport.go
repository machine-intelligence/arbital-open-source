// updateSupport.go can change data for a support.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateSupportTask is the object that's put into the daemon queue.
type updateSupportTask struct {
	Id   int64 `json:",string"`
	Text string
}

// updateSupportHandler renders the support page.
func updateSupportHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateSupportTask
	err := decoder.Decode(&task)
	if err != nil || task.Text == "" {
		c.Inc("update_support_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	updateArgs := make([]string, 0)
	hashmap["id"] = task.Id
	if task.Text != "" {
		hashmap["text"] = task.Text
		updateArgs = append(updateArgs, "text")
	}
	sql := database.GetInsertSql("support", hashmap, updateArgs...)
	if err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_support_fail")
		c.Errorf("Couldn't update support: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
