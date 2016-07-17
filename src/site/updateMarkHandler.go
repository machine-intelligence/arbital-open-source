// updateMarkHandler.go updates an existing mark.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateMarkData contains data given to us in the request.
type updateMarkData struct {
	MarkId string

	// Optional vars.
	// Set mark's text to this
	Text    string
	Submit  bool
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
	db := params.DB
	u := params.U
	loadData := core.NewHandlerData(u)

	var data updateMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Load the mark
	mark := &core.Mark{ID: data.MarkId}
	loadData.MarkMap[data.MarkId] = mark
	err = core.LoadMarkData(db, loadData.PageMap, loadData.UserMap, loadData.MarkMap, u)
	if err != nil {
		return pages.Fail("Couldn't load the mark", err)
	} else if mark.Type == "" {
		return pages.Fail("No such mark", nil).Status(http.StatusBadRequest)
	}

	// Update existing mark
	hashmap := make(database.InsertMap)
	hashmap["id"] = data.MarkId
	if data.Text != "" {
		hashmap["text"] = data.Text
	}
	if data.Submit {
		hashmap["isSubmitted"] = true
	}
	if data.Dismiss {
		hashmap["resolvedBy"] = u.ID
	}
	statement := db.NewInsertStatement("marks", hashmap, hashmap.GetKeys()...)
	_, err = statement.Exec()
	if err != nil {
		return pages.Fail("Couldn't update the mark", err)
	}

	// If the mark has just been submitted, queue the updates
	if data.Submit && !mark.IsSubmitted {
		err = EnqueueNewMarkUpdateTask(params, data.MarkId, mark.PageID, 0)
		if err != nil {
			return pages.Fail("Couldn't enqueue an updateTask", err)
		}
	}

	return pages.Success(nil)
}
