// newMemberHandler.go adds a new member for a group.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// newMemberData contains data given to us in the request.
type newMemberData struct {
	GroupId int64 `json:",string"`
	UserId  int64 `json:",string"`
}

var newMemberHandler = siteHandler{
	URI:         "/newMember/",
	HandlerFunc: newMemberHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newMemberHandlerFunc handles requests to add a new member to a group.
func newMemberHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data newMemberData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.GroupId <= 0 || data.UserId <= 0 {
		return pages.HandlerBadRequestFail("GroupId and UserId have to be set", nil)
	}

	// Check to see if this user can add members.
	var blank int64
	var found bool
	row := db.NewStatement(`
		SELECT 1
		FROM groupMembers
		WHERE userId=? AND groupId=? AND canAddMembers
		`).QueryRow(u.Id, data.GroupId)
	found, err = row.Scan(&blank)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a group member", err)
	} else if !found {
		return pages.HandlerForbiddenFail("You don't have the permission to add a user", nil)
	}

	// Check to see if the proposed member exists.
	row = db.NewStatement(`
		SELECT 1
		FROM users
		WHERE id=?`).QueryRow(data.UserId)
	found, err = row.Scan(&blank)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a user", err)
	} else if !found {
		return pages.HandlerErrorFail("New member id doesn't correspond to a user", nil)
	}

	// Update groupMembers table
	hashmap := make(map[string]interface{})
	hashmap["userId"] = data.UserId
	hashmap["groupId"] = data.GroupId
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("groupMembers", hashmap)
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't add a member", err)
	}

	// Create a task to do further processing
	var task tasks.MemberUpdateTask
	task.UserId = u.Id
	task.UpdateType = core.AddedToGroupUpdateType
	task.MemberId = data.UserId
	task.GroupId = data.GroupId
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	} else if err := tasks.Enqueue(c, task, "memberUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
