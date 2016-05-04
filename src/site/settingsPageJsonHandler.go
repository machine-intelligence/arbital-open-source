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
	DomainId string `json:"domainId"`
	LongName string `json:"longName"`
}

var settingsPageHandler = siteHandler{
	URI:         "/json/settingsPage/",
	HandlerFunc: settingsPageJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
		RequireLogin:    true,
	},
}

// settingsPageJsonHandler renders the settings page.
func settingsPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Get all domains, for user to select when creating invite
	rows := database.NewQuery(`
		SELECT p.pageId,title
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		JOIN pages AS p
		ON (p.pageId=pi.pageId AND p.edit=pi.currentEdit)
		WHERE type=?`, core.DomainPageType).ToStatement(db).Query()
	domains := make(map[string]*Domain)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId, longName string
		err := rows.Scan(&domainId, &longName)
		if err != nil {
			return fmt.Errorf("failed to scan a domain: %v", err)
		}
		domains[domainId] = &Domain{domainId, longName}
		return nil
	})
	if err != nil {
		return pages.HandlerErrorFail("Error while loading domain ids", err)
	}
	returnData.ResultMap["domains"] = domains

	// Get all of the invites a user has SENT
	wherePart := database.NewQuery(`WHERE fromUserId=?`, u.Id)
	returnData.ResultMap["invitesSent"], err = core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load sent invites", err)
	}

	_, err = core.LoadAllDomainIds(db, returnData.PageMap)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load domain ids", err)
	}
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}