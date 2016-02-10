// base10ToBase36Part1Handler.go kicks off the task
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var base10ToBase36Part1Handler = siteHandler{
	URI:         "/base10ToBase36Part1/",
	HandlerFunc: base10ToBase36Part1HandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// base10ToBase36Part1HandlerFunc kicks off the task.
func base10ToBase36Part1HandlerFunc(params *pages.HandlerParams) *pages.Result {
	var task tasks.Base10ToBase36Part1Task
	if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part1"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
