// updateMarkHandler.go updates an existing mark.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// updateMarkData contains data given to us in the request.
type updateMarkData struct {
	MarkId string

	// Optional vars
	// Set mark's text to this
	Text string
	// If true, the user wants to submit this mark to page authors
	Submit bool
	// Resolve the mark, and this is the page id with the "answer"
	ResolvedPageId string
	// Resolve the mark without a page id
	Dismiss bool
}

var updateMarkHandler = siteHandler{
	URI:         "/updateMark/",
	HandlerFunc: updateMarkHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateMarkHandlerFunc handles requests to create/update a prior like.
func updateMarkHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U
	loadData := core.NewHandlerData(u)

	var data updateMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Load the mark
	mark := &core.Mark{Id: data.MarkId}
	loadData.MarkMap[data.MarkId] = mark
	err = core.LoadMarkData(db, loadData.PageMap, loadData.UserMap, loadData.MarkMap, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the mark", err)
	}
	if mark.Type == "" {
		return pages.HandlerBadRequestFail("No such mark", nil)
	}

	// Update existing mark
	hashmap := make(database.InsertMap)
	hashmap["id"] = data.MarkId
	if data.Text != "" {
		hashmap["text"] = data.Text
		hashmap["resolvedPageId"] = ""
		hashmap["resolvedBy"] = ""
	}
	if data.ResolvedPageId != "" {
		hashmap["resolvedPageId"] = data.ResolvedPageId
		hashmap["resolvedBy"] = u.Id
		hashmap["resolvedAt"] = database.Now()
	} else if data.Dismiss {
		hashmap["resolvedBy"] = u.Id
		hashmap["resolvedAt"] = database.Now()
	}
	statement := db.NewInsertStatement("marks", hashmap, hashmap.GetKeys()...)
	_, err = statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't update the mark", err)
	}

	// If the mark was resolved, update the user
	if mark.Type != core.QueryMarkType && mark.ResolvedBy == "" && data.ResolvedPageId != "" {
		hashmap := make(database.InsertMap)
		hashmap["userId"] = mark.CreatorId
		hashmap["type"] = core.ResolvedMarkUpdateType
		hashmap["groupByPageId"] = data.ResolvedPageId
		hashmap["goToPageId"] = data.ResolvedPageId
		hashmap["markId"] = data.MarkId
		hashmap["createdAt"] = database.Now()
		hashmap["unseen"] = true
		statement := db.NewInsertStatement("updates", hashmap)
		_, err = statement.Exec()
		if err != nil {
			return pages.HandlerErrorFail("Couldn't insert update", err)
		}
	}

	// If the user submitted a query mark, notify the authors
	if mark.Type == core.QueryMarkType && mark.Text == "" && data.Text != "" && data.Submit {
		var updateTask tasks.NewUpdateTask
		updateTask.UserId = u.Id
		updateTask.GoToPageId = mark.PageId
		updateTask.SubscribedToId = mark.PageId
		updateTask.UpdateType = core.NewMarkUpdateType
		updateTask.GroupByPageId = mark.PageId
		updateTask.MarkId = mark.Id
		updateTask.EditorsOnly = true
		if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
			return pages.HandlerErrorFail("Couldn't enqueue an updateTask", err)
		}
	}

	return pages.StatusOK(nil)
}
