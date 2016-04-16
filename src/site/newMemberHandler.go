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
	GroupId   string
	UserInput string
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
	if !core.IsIdValid(data.GroupId) {
		return pages.HandlerBadRequestFail("GroupId has to be set", nil)
	}
	if data.UserInput == "" {
		return pages.HandlerBadRequestFail("Must provide an identifier for the new user", nil)
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

	var newMemberId string

	// See if data.UserInput is the id or email of an existing user
	row = db.NewStatement(`
		SELECT id
		FROM users
		WHERE id=? OR email=?`).QueryRow(data.UserInput, data.UserInput)
	found, err = row.Scan(&newMemberId)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a user", err)
	}

	if !found {
		// See if data.UserInput is the alias of the user's page
		row = db.NewStatement(`
			SELECT pageId
			FROM pageInfos
			WHERE alias=?`).QueryRow(data.UserInput)
		// The id of this page is the same as the id of the user we want
		found, err = row.Scan(&newMemberId)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't check for a user", err)
		} else if !found {
			return pages.HandlerErrorFail("Couldn't find the user", nil)
		}
	}

	// Update groupMembers table
	hashmap := make(map[string]interface{})
	hashmap["userId"] = newMemberId
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
	task.MemberId = newMemberId
	task.GroupId = data.GroupId
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
