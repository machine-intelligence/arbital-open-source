// groupsJsonHandler.go returns the data about user's groups.

package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var groupsHandler = siteHandler{
	URI:         "/json/groups/",
	HandlerFunc: groupsJSONHandler,
	Options:     pages.PageOptions{},
}

func groupsJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Load the groups and members
	rows := database.NewQuery(`
		SELECT p.pageId,m.userId,m.canAddMembers,m.canAdmin
		FROM pages AS p
		LEFT JOIN (
			SELECT userId,groupId,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (p.pageId=m.groupId)
		WHERE p.pageId IN`).AddArgsGroupStr(u.GroupIds).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var groupID string
		var m core.Member
		err := rows.Scan(&groupID, &m.UserID, &m.CanAddMembers, &m.CanAdmin)
		if err != nil {
			return fmt.Errorf("failed to scan a group member: %v", err)
		}

		// Add group
		curGroup := core.AddPageIDToMap(groupID, returnData.PageMap)
		if curGroup.Members == nil {
			curGroup.Members = make(map[string]*core.Member)
		}

		// Add member
		curGroup.Members[m.UserID] = &m
		returnData.UserMap[m.UserID] = &core.User{ID: m.UserID}
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading group members", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
