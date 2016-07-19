// newMemberHandler.go adds a new member for a group.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// newMemberData contains data given to us in the request.
type newMemberData struct {
	GroupID   string
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.GroupID) {
		return pages.Fail("GroupId has to be set", nil).Status(http.StatusBadRequest)
	}
	if data.UserInput == "" {
		return pages.Fail("Must provide an identifier for the new user", nil).Status(http.StatusBadRequest)
	}

	// Check to see if this user can add members.
	var blank int64
	var found bool
	row := db.NewStatement(`
		SELECT 1
		FROM groupMembers
		WHERE userId=? AND groupId=? AND canAddMembers
		`).QueryRow(u.ID, data.GroupID)
	found, err = row.Scan(&blank)
	if err != nil {
		return pages.Fail("Couldn't check for a group member", err)
	} else if !found {
		return pages.Fail("You don't have the permission to add a user", nil).Status(http.StatusForbidden)
	}

	var newMemberID string

	// See if data.UserInput is the id or email of an existing user
	row = db.NewStatement(`
		SELECT id
		FROM users
		WHERE id=? OR email=?`).QueryRow(data.UserInput, data.UserInput)
	found, err = row.Scan(&newMemberID)
	if err != nil {
		return pages.Fail("Couldn't check for a user", err)
	}

	if !found {
		// See if data.UserInput is the alias of the user's page
		// The id of this page is the same as the id of the user we want
		found, err = database.NewQuery(`
			SELECT pageId
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			WHERE alias=?`, data.UserInput).ToStatement(db).QueryRow().Scan(&newMemberID)
		if err != nil {
			return pages.Fail("Couldn't check for a user", err)
		} else if !found {
			return pages.Fail("Couldn't find the user", nil)
		}
	}

	// Update groupMembers table
	hashmap := make(map[string]interface{})
	hashmap["userId"] = newMemberID
	hashmap["groupId"] = data.GroupID
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("groupMembers", hashmap)
	if _, err = statement.Exec(); err != nil {
		return pages.Fail("Couldn't add a member", err)
	}

	// Create a task to do further processing
	var task tasks.MemberUpdateTask
	task.UserID = u.ID
	task.UpdateType = core.AddedToGroupUpdateType
	task.MemberID = newMemberID
	task.GroupID = data.GroupID
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
