// fixTextHandler.go kicks off the task
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var fixTextHandler = siteHandler{
	URI:         "/fixText/",
	HandlerFunc: fixTextHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// fixTextHandlerFunc kicks off the task.
func fixTextHandlerFunc(params *pages.HandlerParams) *pages.Result {
	var task tasks.FixTextTask
	if err := task.IsValid(); err != nil {
		return pages.HandlerErrorFail("Invalid task created", err)
	} else if err := tasks.Enqueue(params.C, task, "fixText"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
