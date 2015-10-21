// deletePagePairHandler.go handles requests for deleting a tag.
package site

import (
	"encoding/json"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// deletePagePairData contains the data we receive in the request.
type deletePagePairData struct {
	ParentId int64 `json:",string"`
	ChildId  int64 `json:",string"`
	Type     string
}

// deletePagePairHandler handles requests for deleting a tag.
func deletePagePairHandler(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	decoder := json.NewDecoder(params.R.Body)
	var data deletePagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.ParentId <= 0 || data.ChildId <= 0 {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", err)
	}
	data.Type = strings.ToLower(data.Type)
	if data.Type != core.ParentPagePairType &&
		data.Type != core.TagPagePairType &&
		data.Type != core.RequirementPagePairType {
		return pages.HandlerBadRequestFail("Incorrect type", err)
	}

	// Delete the pair
	query := db.NewStatement(`
		DELETE FROM pagePairs
		WHERE parentId=? AND childId=? AND type=?`)
	if _, err := query.Exec(data.ParentId, data.ChildId, data.Type); err != nil {
		return pages.HandlerErrorFail("Couldn't delete a tag", err)
	}

	if data.Type == core.ParentPagePairType || data.Type == core.TagPagePairType {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.ChildId
		if err := task.IsValid(); err != nil {
			c.Errorf("Invalid task created: %v", err)
		} else if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}
	return pages.StatusOK(nil)
}
