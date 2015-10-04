// fixTextHandler.go kicks off the task
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// fixTextHandler kicks off the task.
func fixTextHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	if !u.IsAdmin {
		return pages.HandlerForbiddenFail("Have to be an admin", nil)
	}

	var task tasks.FixTextTask
	if err := task.IsValid(); err != nil {
		return pages.HandlerErrorFail("Invalid task created", err)
	} else if err := tasks.Enqueue(params.C, task, "fixText"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
