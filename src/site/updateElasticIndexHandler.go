// updateElasticIndexHandler.go kicks off the task to update the index for pages.
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var updateElasticIndexHandler = siteHandler{
	URI:         "/updateElasticIndex/",
	HandlerFunc: updateElasticIndexHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// updateElasticIndexHandlerFunc kicks off the task to update the index for pages.
func updateElasticIndexHandlerFunc(params *pages.HandlerParams) *pages.Result {
	var task tasks.PopulateElasticTask
	if err := tasks.Enqueue(params.C, &task, "populateElastic"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task", err)
	}
	return pages.StatusOK(nil)
}
