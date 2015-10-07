// updateElasticIndexHandler.go kicks off the task to update the index for pages.
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// updateElasticIndexHandler kicks off the task to update the index for pages.
func updateElasticIndexHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U

	if !u.IsAdmin {
		return pages.HandlerForbiddenFail("Have to be an admin", nil)
	}

	var task tasks.PopulateElasticTask
	if err := task.IsValid(); err != nil {
		return pages.HandlerErrorFail("Invalid task created", err)
	} else if err := tasks.Enqueue(params.C, task, "populateElastic"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
