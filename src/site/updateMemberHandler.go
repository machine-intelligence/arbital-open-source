// updateMemberHandler.go adds a new member for a group.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// updateMemberData contains data given to us in the request.
type updateMemberData struct {
	GroupID       string
	UserID        string
	CanAddMembers bool
	CanAdmin      bool
}

var updateMemberHandler = siteHandler{
	URI:         "/updateMember/",
	HandlerFunc: updateMemberHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateMemberHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateMemberData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.GroupID) || !core.IsIDValid(data.UserID) {
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
		return pages.Fail("You don't have the permission to add a user", nil).Status(http.StatusForbidden)
	}

	// Check if the target user exists and get their permissions
	var targetCanAdmin bool
	row = db.NewStatement(`
		SELECT canAdmin
		FROM groupMembers
		WHERE userId=? AND groupId=?
		`).QueryRow(data.UserID, data.GroupID)
	found, err = row.Scan(&targetCanAdmin)
	if err != nil {
		return pages.Fail("Couldn't check for target group member", err)
	} else if !found {
		return pages.Fail("Target member not found", nil).Status(http.StatusForbidden)
	}

	// Admin's can't change property on non-admin.
	if !canAdmin && targetCanAdmin {
		data.CanAdmin = targetCanAdmin
	}
	data.CanAddMembers = data.CanAddMembers || data.CanAdmin

	hashmap := make(map[string]interface{})
	hashmap["userId"] = data.UserID
	hashmap["groupId"] = data.GroupID
	hashmap["canAddMembers"] = data.CanAddMembers
	hashmap["canAdmin"] = data.CanAdmin
	statement := db.NewReplaceStatement("groupMembers", hashmap)
	if _, err = statement.Exec(); err != nil {
		return pages.Fail("Couldn't update a member", err)
	}
	return pages.Success(nil)
}
