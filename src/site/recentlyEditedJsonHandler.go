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

const RecentlyEditedIdsHandlerType = "recentlyEditedIds"

func recentlyEditedJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadRecentlyEdited, RecentlyEditedIdsHandlerType)
}

func LoadRecentlyEdited(db *database.DB, returnData *core.CommonHandlerData, privateDomainID string, numToLoad int,
	pageOptions *core.PageLoadOptions) ([]string, error) {
	// Load recently created and edited by me page ids
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		WHERE p.creatorId=?`, returnData.User.ID).Add(`
			AND pi.seeDomainId=?`, privateDomainID).Add(`
			AND pi.type!=?`, core.CommentPageType).Add(`
			AND`).AddPart(core.WherePageInfos(returnData.User)).Add(`
		GROUP BY 1
		ORDER BY MAX(p.createdAt) DESC
		LIMIT ?`, numToLoad).ToStatement(db).Query()
	return core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
}
