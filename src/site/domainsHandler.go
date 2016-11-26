// domainsHandler.go returns the data about current user's domains.

package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var domainsHandler = siteHandler{
	URI:         "/json/domains/",
	HandlerFunc: domainsHandlerFunc,
	Options:     pages.PageOptions{},
}

func domainsHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	currentUserDomainIDs := make([]string, 0)
	for domainID := range u.DomainMembershipMap {
		currentUserDomainIDs = append(currentUserDomainIDs, domainID)
	}

	// Load all members
	rows := database.NewQuery(`
		SELECT domainId,userId,createdAt,role
		FROM domainMembers
		WHERE domainId IN`).AddArgsGroupStr(currentUserDomainIDs).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var dm core.DomainMember
		err := rows.Scan(&dm.DomainID, &dm.UserID, &dm.CreatedAt, &dm.Role)
		if err != nil {
			return fmt.Errorf("failed to scan for a member: %v", err)
		}
		user := core.AddUserToMap(dm.UserID, returnData.UserMap)
		user.DomainMembershipMap[dm.DomainID] = &dm
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading domain members", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
