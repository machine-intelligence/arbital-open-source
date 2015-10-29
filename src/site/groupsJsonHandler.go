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

	// Load user groups
	err := core.LoadUserGroupIds(db, u)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load user groups", err)
	}

	// Load the groups and members
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	rows := database.NewQuery(`
		SELECT p.pageId,m.userId,m.canAddMembers,m.canAdmin
		FROM pages AS p
		LEFT JOIN (
			SELECT userId,groupId,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (p.pageId=m.groupId)
		WHERE p.pageId IN`).AddArgsGroupStr(u.GroupIds).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var groupId int64
		var m core.Member
		err := rows.Scan(&groupId, &m.UserId, &m.CanAddMembers, &m.CanAdmin)
		if err != nil {
			return fmt.Errorf("failed to scan a group member: %v", err)
		}

		// Add group
		curGroup := core.AddPageIdToMap(groupId, pageMap)
		if curGroup.Members == nil {
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

	// Load pages.
	core.AddUserGroupIdsToPageMap(u, pageMap)
	err = core.LoadPages(db, pageMap, u.Id, &core.LoadPageOptions{LoadSummary: true})
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Load aux data
	err = core.LoadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return pages.Fail("error while loading aux data", err)
	}

	// Load all the users.
	err = core.LoadUsers(db, userMap)
	if err != nil {
		return pages.Fail("Error while loading users", err)
	}

	returnData := createReturnData(pageMap).AddUsers(userMap)
	return pages.StatusOK(returnData)
}
