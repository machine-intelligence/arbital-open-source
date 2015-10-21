// newGroupHandler.go creates a new group.
package site

import (
	"encoding/json"
	"math/rand"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// newGroupData contains data given to us in the request.
type newGroupData struct {
	Name string

	IsDomain   bool
	Alias      string
	RootPageId int64 `json:",string"`
}

// newGroupHandler handles requests to add a new group to a group.
func newGroupHandler(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data newGroupData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.Name == "" {
		return pages.HandlerBadRequestFail("Name has to be set", nil)
	}

	// Check user related permissions.
	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Not logged in", nil)
	}
	if u.Karma < 200 {
		return pages.HandlerForbiddenFail("You don't have enough karma", nil)
	}
	if !u.IsAdmin {
		return pages.HandlerForbiddenFail("Have to be an admin to create domains or groups", nil)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		groupId := rand.Int63()

		// Create the new group.
		hashmap := make(map[string]interface{})
		hashmap["id"] = groupId
		hashmap["name"] = data.Name
		hashmap["createdAt"] = database.Now()
		if data.IsDomain {
			hashmap["isDomain"] = true
			hashmap["alias"] = data.Alias
			hashmap["rootPageId"] = data.RootPageId
		}
		statement := tx.NewInsertTxStatement("groups", hashmap)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't create a group", err
		}

		if !data.IsDomain {
			// Add the user to the group as an admin.
			hashmap = make(map[string]interface{})
			hashmap["userId"] = u.Id
			hashmap["groupId"] = groupId
			hashmap["canAddMembers"] = true
			hashmap["canAdmin"] = true
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("groupMembers", hashmap)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add a user to a group", err
			}
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	if data.IsDomain {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.RootPageId
		if err := task.IsValid(); err != nil {
			c.Errorf("Invalid task created: %v", err)
		}
		if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return pages.StatusOK(nil)
}
