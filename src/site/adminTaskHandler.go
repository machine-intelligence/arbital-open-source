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
	task := params.R.FormValue("task")
	if task == "fixText" {
		var task tasks.FixTextTask
		if err := tasks.Enqueue(params.C, &task, "fixText"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "populateElastic" {
		var task tasks.PopulateElasticTask
		if err := tasks.Enqueue(params.C, &task, "populateElastic"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "updateMetadata" {
		var task tasks.UpdateMetadataTask
		if err := tasks.Enqueue(params.C, &task, "updateMetadata"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "propagateDomain" {
		var task tasks.PropagateDomainTask
		task.PageId = params.R.FormValue("pageId")
		if err := tasks.Enqueue(params.C, &task, "propagateDomain"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "resetPasswords" {
		var task tasks.ResetPasswordsTask
		if err := tasks.Enqueue(params.C, &task, "resetPasswords"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "base10ToBase36Part1" {
		var task tasks.Base10ToBase36Part1Task
		if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part1"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "base10ToBase36Part2" {
		var task tasks.Base10ToBase36Part2Task
		if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part2"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "base10ToBase36Part3" {
		var task tasks.Base10ToBase36Part3Task
		if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part3"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "base10ToBase36Part4" {
		var task tasks.Base10ToBase36Part4Task
		if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part4"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else if task == "base10ToBase36Part5" {
		var task tasks.Base10ToBase36Part5Task
		if err := tasks.Enqueue(params.C, &task, "base10ToBase36Part5"); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue a task", err)
		}
	} else {
		return pages.HandlerErrorFail("Unknown ?task var", nil)
	}
	return pages.StatusOK(nil)
}
