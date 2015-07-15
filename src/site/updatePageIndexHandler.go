// updatePageIndexHandler.go kicks off the task to update the index for pages.
package site

import (
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// updatePageIndexHandler kicks off the task to update the index for pages.
func updatePageIndexHandler(w http.ResponseWriter, r *http.Request) {
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

	var task tasks.PopulateIndexTask
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "populateIndex"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
