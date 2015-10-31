// deleteMemberHandler.go deletes a group member from the group
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// deleteMemberData contains data given to us in the request.
type deleteMemberData struct {
	GroupId int64 `json:",string"`
	UserId  int64 `json:",string"`
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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.GroupId <= 0 || data.UserId <= 0 {
		return pages.HandlerBadRequestFail("GroupId and UserId have to be set", nil)
	}

	// Check to see if this user can add members.
	var canAdmin bool
	row := db.NewStatement(`
		SELECT canAdmin
		FROM groupMembers
		WHERE userId=? AND groupId=? AND canAddMembers
		`).QueryRow(u.Id, data.GroupId)
	found, err := row.Scan(&canAdmin)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a group member", err)
	} else if !found {
		return pages.HandlerForbiddenFail("You don't have the permission to remove a user", nil)
	}

	// Check if the target user exists and get their permissions
	var targetCanAdmin bool
	row = db.NewStatement(`
		SELECT canAdmin
		FROM groupMembers
		WHERE userId=? AND groupId=?
		`).QueryRow(data.UserId, data.GroupId)
	found, err = row.Scan(&targetCanAdmin)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for target group member", err)
	} else if !found {
		return pages.HandlerForbiddenFail("Target member not found", nil)
	}

	// Non-admins can't delete admins.
	if !canAdmin && targetCanAdmin {
		return pages.HandlerForbiddenFail("Non-admins can't remove admins", nil)
	}

	// Check to see if the proposed member exists.
	statement := db.NewStatement(`
		DELETE FROM groupMembers
		WHERE userId=? AND groupId=?`)
	if _, err := statement.Exec(data.UserId, data.GroupId); err != nil {
		return pages.HandlerErrorFail("Couldn't delete the group member", err)
	}

	// Create a task to do further processing
	var task tasks.MemberUpdateTask
	task.UserId = u.Id
	task.UpdateType = core.RemovedFromGroupUpdateType
	task.MemberId = data.UserId
	task.GroupId = data.GroupId
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	} else if err := tasks.Enqueue(c, task, "memberUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
