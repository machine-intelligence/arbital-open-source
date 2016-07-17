// deleteMemberHandler.go deletes a group member from the group
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// deleteMemberData contains data given to us in the request.
type deleteMemberData struct {
	GroupID string
	UserId  string
}

var deleteMemberHandler = siteHandler{
	URI:         "/deleteMember/",
	HandlerFunc: deleteMemberHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func deleteMemberHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	var data deleteMemberData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.GroupID) || !core.IsIdValid(data.UserId) {
		return pages.Fail("GroupId and UserId have to be set", nil).Status(http.StatusBadRequest)
	}

	// Check to see if this user can add members.
	var canAdmin bool
	row := db.NewStatement(`
		SELECT canAdmin
		FROM groupMembers
		WHERE userId=? AND groupId=? AND canAddMembers
		`).QueryRow(u.ID, data.GroupID)
	found, err := row.Scan(&canAdmin)
	if err != nil {
		return pages.Fail("Couldn't check for a group member", err)
	} else if !found {
		return pages.Fail("You don't have the permission to remove a user", nil).Status(http.StatusForbidden)
	}

	// Check if the target user exists and get their permissions
	var targetCanAdmin bool
	row = db.NewStatement(`
		SELECT canAdmin
		FROM groupMembers
		WHERE userId=? AND groupId=?
		`).QueryRow(data.UserId, data.GroupID)
	found, err = row.Scan(&targetCanAdmin)
	if err != nil {
		return pages.Fail("Couldn't check for target group member", err)
	} else if !found {
		return pages.Fail("Target member not found", nil).Status(http.StatusForbidden)
	}

	// Non-admins can't delete admins.
	if !canAdmin && targetCanAdmin {
		return pages.Fail("Non-admins can't remove admins", nil).Status(http.StatusForbidden)
	}

	// Check to see if the proposed member exists.
	statement := db.NewStatement(`
		DELETE FROM groupMembers
		WHERE userId=? AND groupId=?`)
	if _, err := statement.Exec(data.UserId, data.GroupID); err != nil {
		return pages.Fail("Couldn't delete the group member", err)
	}

	// Create a task to do further processing
	var task tasks.MemberUpdateTask
	task.UserId = u.ID
	task.UpdateType = core.RemovedFromGroupUpdateType
	task.MemberId = data.UserId
	task.GroupID = data.GroupID
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
