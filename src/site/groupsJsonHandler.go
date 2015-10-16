// groupsJsonHandler.go returns the data about user's groups.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

func groupsJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Load the groups and members
	userMap := make(map[int64]*core.User)
	groupMap := make(map[int64]*core.Group)
	rows := db.NewStatement(`
		SELECT g.id,g.name,m.userId,m.canAddMembers,m.canAdmin
		FROM groups AS g
		LEFT JOIN (
			SELECT userId,groupId,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (g.id=m.groupId)
		WHERE g.id IN (SELECT groupId FROM groupMembers WHERE userId=?)`).Query(u.Id)
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
		curGroup := groupMap[g.Id]
		if curGroup == nil {
			curGroup = &g
			groupMap[g.Id] = curGroup
			curGroup.Members = make(map[string]*core.Member)
		}

		// Add member
		curGroup.Members[fmt.Sprintf("%d", m.UserId)] = &m
		userMap[m.UserId] = &core.User{Id: m.UserId}
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading group members", err)
	}

	// Load all the users.
	err = core.LoadUsers(db, userMap)
	if err != nil {
		return pages.Fail("Error while loading users", err)
	}

	returnData := createReturnData(nil).AddUsers(userMap).AddGroups(groupMap)
	return pages.StatusOK(returnData)
}
