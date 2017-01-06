// Serves JSON for most recently created comments

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var recentlyCreatedCommentHandler = siteHandler{
	URI:         "/json/recentlyCreatedComment/",
	HandlerFunc: recentlyCreatedCommentJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

const RecentlyCreatedCommentIdsHandlerType = "recentlyCreatedCommentIds"

func recentlyCreatedCommentJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadRecentlyCreatedComment, RecentlyCreatedCommentIdsHandlerType)
}

func LoadRecentlyCreatedComment(db *database.DB, returnData *core.CommonHandlerData, privateDomainID string, numToLoad int,
	_ *core.PageLoadOptions) ([]string, error) {
	// Load recently created by me comment ids
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=?`, returnData.User.ID).Add(`
			AND pi.seeDomainId=?`, privateDomainID).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
			AND`).AddPart(core.PageInfosFilter(returnData.User)).Add(`
		ORDER BY pi.createdAt DESC
		LIMIT ?`, numToLoad).ToStatement(db).Query()
	return core.LoadPageIDs(rows, returnData.PageMap, core.TitlePlusLoadOptions)
}
