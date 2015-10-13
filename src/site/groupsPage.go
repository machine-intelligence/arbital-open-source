// groupsPage.go serves the groups template.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// groupsTmplData stores the data that we pass to the template to render the page
type groupsTmplData struct {
	commonPageData
}

// groupsPage serves the recent pages page.
var groupsPage = newPage(
	"/groups/",
	groupsRenderer,
	append(baseTmpls,
		"tmpl/groupsPage.tmpl", "tmpl/angular.tmpl.js"))

// groupsRenderer renders the page page.
func groupsRenderer(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data groupsTmplData
	data.User = u

	// Load the groups and members
	data.UserMap = make(map[int64]*core.User)
	data.GroupMap = make(map[int64]*core.Group)
	rows := db.NewStatement(`
		SELECT g.id,g.name,m.userId,m.canAddMembers,m.canAdmin
		FROM groups AS g
		LEFT JOIN (
			SELECT userId,groupId,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (g.id=m.groupId)
		WHERE g.id IN (SELECT groupId FROM groupMembers WHERE userId=?)`).Query(data.User.Id)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var g core.Group
		var m core.Member
		err := rows.Scan(
			&g.Id,
			&g.Name,
			&m.UserId,
			&m.CanAddMembers,
			&m.CanAdmin)
		if err != nil {
			return fmt.Errorf("failed to scan a group member: %v", err)
		}

		// Add group
		curGroup := data.GroupMap[g.Id]
		if curGroup == nil {
			curGroup = &g
			data.GroupMap[g.Id] = curGroup
		}

		// Add member
		curGroup.Members = append(curGroup.Members, &m)
		if m.UserId == data.User.Id {
			curGroup.UserMember = &m
		}
		data.UserMap[m.UserId] = &core.User{Id: m.UserId}
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading group members", err)
	}

	// Load all the users.
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return pages.Fail("Error while loading users", err)
	}

	return pages.StatusOK(data)
}
