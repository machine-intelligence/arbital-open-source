// updateMetadataHandler.go kicks off the task to update metadata for all pages
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var updateMetadataHandler = siteHandler{
	URI:         "/updateMetadata/",
	HandlerFunc: updateMetadataHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// updateMetadataHandlerFunc kicks off the task to update the index for pages.
func updateMetadataHandlerFunc(params *pages.HandlerParams) *pages.Result {
	var task tasks.UpdateMetadataTask
	if err := task.IsValid(); err != nil {
		return pages.HandlerErrorFail("Invalid task created", err)
	} else if err := tasks.Enqueue(params.C, task, "updateMetadata"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
