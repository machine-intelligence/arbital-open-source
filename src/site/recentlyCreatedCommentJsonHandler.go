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

const RecentlyCreatedCommentIds = "recentlyCreatedCommentIds"

func recentlyCreatedCommentJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadRecentlyCreatedComment, RecentlyCreatedCommentIds)
}

func LoadRecentlyCreatedComment(u *core.CurrentUser, privateGroupID string, numToLoad int, db *database.DB, returnData *core.CommonHandlerData,
	_ *core.PageLoadOptions) ([]string, error) {
	// Load recently created by me comment ids
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=?`, u.ID).Add(`
			AND pi.seeGroupId=?`, privateGroupID).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
		ORDER BY pi.createdAt DESC
		LIMIT ?`, numToLoad).ToStatement(db).Query()
	return core.LoadPageIDs(rows, returnData.PageMap, core.TitlePlusLoadOptions)
}
