// newGroupHandler.go creates a new group.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// newGroupData contains data given to us in the request.
type newGroupData struct {
	Name string

	IsDomain   bool
	Alias      string
	RootPageID string
}

var newGroupHandler = siteHandler{
	URI:         "/newGroup/",
	HandlerFunc: newGroupHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		AdminOnly:    true,
	},
}

// newGroupHandlerFunc handles requests to add a new group to a group.
func newGroupHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	var data newGroupData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if data.Name == "" {
		return pages.Fail("Name has to be set", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		groupID, err := core.GetNextAvailableID(tx)
		if err != nil {
			return sessions.NewError("Couldn't get next available Id", err)
		}
		if data.IsDomain {
			return core.NewDomain(tx, groupID, u.ID, data.Name, data.Alias)
		}
		return core.NewGroup(tx, groupID, u.ID, data.Name, data.Alias)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	if data.IsDomain {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageID = data.RootPageID
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.Success(nil)
}
