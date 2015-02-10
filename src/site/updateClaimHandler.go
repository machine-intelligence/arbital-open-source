// updateClaim.go can change data for a claim.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateClaimTask is the object that's put into the daemon queue.
type updateClaimTask struct {
	Id      int64 `json:",string"`
	Summary string
	Text    string
	Url     string
}

// updateClaimHandler renders the claim page.
func updateClaimHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateClaimTask
	err := decoder.Decode(&task)
	if err != nil || task.Id <= 0 || task.Text == "" {
		c.Inc("update_claim_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = task.Id
	hashmap["summary"] = task.Summary
	hashmap["text"] = task.Text
	hashmap["url"] = task.Url
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("claims", hashmap, "text", "summary", "url", "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_claim_fail")
		c.Errorf("Couldn't update claim: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
