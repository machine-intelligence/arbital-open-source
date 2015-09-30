// updateMetadataHandler.go kicks off the task to update metadata for all pages
package site

import (
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// updateMetadataHandler kicks off the task to update the index for pages.
func updateMetadataHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	db, err := database.GetDB(c)
	if err != nil {
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	u, err := user.LoadUser(w, r, db)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var task tasks.UpdateMetadataTask
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "updateMetadata"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
