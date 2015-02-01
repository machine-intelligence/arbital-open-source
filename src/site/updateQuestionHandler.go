// updateQuestion.go can change data for a question.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// updateQuestionTask is the object that's put into the daemon queue.
type updateQuestionTask struct {
	Id   int64 `json:",string"`
	Text string
}

// updateQuestionHandler renders the question page.
func updateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var task updateQuestionTask
	err := decoder.Decode(&task)
	if err != nil || task.Id <= 0 || task.Text == "" {
		c.Inc("update_question_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["id"] = task.Id
	hashmap["text"] = task.Text
	hashmap["updatedAt"] = database.Now()
	sql := database.GetInsertSql("questions", hashmap, "text", "updatedAt")
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("update_question_fail")
		c.Errorf("Couldn't update question: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
