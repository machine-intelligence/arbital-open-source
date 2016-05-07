// adminTaskHandler.go kicks off the task
package site

import (
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

var adminTaskHandler = siteHandler{
	URI:         "/adminTask/",
	HandlerFunc: adminTaskHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

// adminTaskHandlerFunc kicks off the task.
func adminTaskHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	task := params.R.FormValue("task")
	if task == "fixText" {
		var task tasks.FixTextTask
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "populateElastic" {
		var task tasks.PopulateElasticTask
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "updateMetadata" {
		var task tasks.UpdateMetadataTask
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "propagateDomain" {
		var task tasks.PropagateDomainTask
		task.PageId = params.R.FormValue("pageId")
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "resetPasswords" {
		var task tasks.ResetPasswordsTask
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "sendInvite" {
		var task tasks.SendInviteTask
		task.FromUserId = "1"
		task.ToEmail = "alexei.andreev@gmail.com"
		task.DomainIds = []string{"1lw", "2v", "3d"}
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else {
		return pages.HandlerErrorFail("Unknown ?task var", nil)
	}
	return pages.StatusOK(nil)
}
