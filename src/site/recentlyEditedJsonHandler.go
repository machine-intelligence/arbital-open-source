// Serves JSON for recently edited pages

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var recentlyEditedHandler = siteHandler{
	URI:         "/json/recentlyEdited/",
	HandlerFunc: recentlyEditedJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

const RecentlyEditedIds = "recentlyEditedIds"

func recentlyEditedJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadRecentlyEdited, RecentlyEditedIds)
}

func LoadRecentlyEdited(u *core.CurrentUser, privateGroupID string, numToLoad int, db *database.DB, returnData *core.CommonHandlerData,
	pageOptions *core.PageLoadOptions) ([]string, error) {
	// Load recently created and edited by me page ids
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId)
		WHERE p.creatorId=?`, u.ID).Add(`
			AND pi.seeGroupId=?`, privateGroupID).Add(`
			AND pi.type!=?`, core.CommentPageType).Add(`
		GROUP BY 1
		ORDER BY MAX(p.createdAt) DESC
		LIMIT ?`, numToLoad).ToStatement(db).Query()
	return core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
}
