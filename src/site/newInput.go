// newInput.go handles requests to new a new input to the database.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// newInputData is the object that's put into the daemon queue.
type newInputData struct {
	QuestionId int64 `json:",string"`
	Text       string
}

func newInputHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newInputData
	err := decoder.Decode(&data)
	if err != nil || data.Text == "" || data.QuestionId <= 0 {
		c.Inc("new_input_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_input_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create new input.
	hashmap := make(map[string]interface{})
	hashmap["questionId"] = data.QuestionId
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["createdAt"] = database.Now()
	hashmap["text"] = data.Text
	sql := database.GetInsertSql("inputs", hashmap)
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("new_input_fail")
		c.Errorf("Couldn't new input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add updates to users who are subscribed to this question.
	var task tasks.NewUpdateTask
	task.UserId = u.Id
	task.QuestionId = data.QuestionId
	task.UpdateType = "newInput"
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
