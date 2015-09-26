// updateElasticIndexHandler.go kicks off the task to update the index for pages.
package site

import (
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// updateElasticIndexHandler kicks off the task to update the index for pages.
func updateElasticIndexHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Get user object
	u, err := user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var task tasks.PopulateElasticTask
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "populateElastic"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
