// settingsPageJsonHandler.go contains the handler for returning JSON with data
// to display the settings/invite page.

package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type Domain struct {
	DomainID string `json:"domainId"`
	LongName string `json:"longName"`
}

var settingsPageHandler = siteHandler{
	URI:         "/json/settingsPage/",
	HandlerFunc: settingsPageJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// settingsPageJsonHandler renders the settings page.
func settingsPageJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Get all domains, for user to select when creating invite
	rows := database.NewQuery(`
		SELECT p.pageId,title
		FROM pages AS p
		JOIN domains AS d
		ON (p.pageId=d.pageId AND p.isLiveEdit)`).ToStatement(db).Query()
	domains := make(map[string]*Domain)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainID, longName string
		err := rows.Scan(&domainID, &longName)
		if err != nil {
			return fmt.Errorf("failed to scan a domain: %v", err)
		}
		domains[domainID] = &Domain{domainID, longName}
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading domain ids", err)
	}
	returnData.ResultMap["domains"] = domains

	// Get all of the invites a user has SENT
	wherePart := database.NewQuery(`WHERE fromUserId=?`, u.ID)
	returnData.ResultMap["invitesSent"], err = core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.Fail("Couldn't load sent invites", err)
	}

	/*_, err = core.LoadAllDomainIDs(db, returnData.PageMap)
	if err != nil {
		return pages.Fail("Couldn't load domain ids", err)
	}*/
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
