// base10ToBase36Part2Handler.go kicks off the task
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var base10ToBase36Part2Handler = siteHandler{
	URI:         "/base10ToBase36Part2/",
	HandlerFunc: base10ToBase36Part2HandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// base10ToBase36Part2HandlerFunc kicks off the task.
func base10ToBase36Part2HandlerFunc(params *pages.HandlerParams) *pages.Result {
	var task tasks.Base10ToBase36Part2Task
	if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part2"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
