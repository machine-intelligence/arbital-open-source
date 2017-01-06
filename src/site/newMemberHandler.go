// newMemberHandler.go adds a new member to a domain.

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
	DomainID  string
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
	if !core.IsIntIDValid(data.DomainID) {
		return pages.Fail("GroupId has to be set", nil).Status(http.StatusBadRequest)
	}
	if data.UserInput == "" {
		return pages.Fail("Must provide an identifier for the new user", nil).Status(http.StatusBadRequest)
	}

	if !core.CanCurrentUserGiveRole(u, data.DomainID, core.DefaultDomainRole) {
		return pages.Fail("You don't have permissions to add users to this domain", nil)
	}

	var newMemberID string

	// See if data.UserInput is the id or email of an existing user
	row := db.NewStatement(`
		SELECT id
		FROM users
		WHERE id=? OR email=?`).QueryRow(data.UserInput, data.UserInput)
	found, err := row.Scan(&newMemberID)
	if err != nil {
		return pages.Fail("Couldn't check for a user", err)
	}

	if !found {
		// See if data.UserInput is the alias of the user's page
		// The id of this page is the same as the id of the user we want
		found, err = database.NewQuery(`
			SELECT pi.pageId
			FROM pageInfos AS pi
			WHERE pi.alias=?`, data.UserInput).Add(`
				AND`).AddPart(core.WherePageInfos(u)).ToStatement(db).QueryRow().Scan(&newMemberID)
		if err != nil {
			return pages.Fail("Couldn't check for a user", err)
		} else if !found {
			return pages.Fail("Couldn't find the user", nil)
		}
	}

	// Update groupMembers table
	hashmap := make(map[string]interface{})
	hashmap["userId"] = newMemberID
	hashmap["domainId"] = data.DomainID
	hashmap["createdAt"] = database.Now()
	hashmap["role"] = string(core.DefaultDomainRole)
	statement := db.NewInsertStatement("domainMembers", hashmap)
	if _, err = statement.Exec(); err != nil {
		return pages.Fail("Couldn't add a member", err)
	}

	// Create a task to do further processing
	var task tasks.MemberUpdateTask
	task.UserID = u.ID
	task.UpdateType = core.AddedToGroupUpdateType
	task.MemberID = newMemberID
	task.DomainID = data.DomainID
	if err := tasks.Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
