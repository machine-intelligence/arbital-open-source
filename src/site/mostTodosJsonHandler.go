// Serves JSON for pages with the most todos

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var mostTodosHandler = siteHandler{
	URI:         "/json/mostTodos/",
	HandlerFunc: mostTodosJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

const MostTodosIdsHandlerType = "mostTodosIds"

func mostTodosJSONHandler(params *pages.HandlerParams) *pages.Result {
	return DashboardListJSONHandler(params, LoadMostTodos, MostTodosIdsHandlerType)
}

func LoadMostTodos(db *database.DB, returnData *core.CommonHandlerData, privateDomainID string, numToLoad int,
	pageOptions *core.PageLoadOptions) ([]string, error) {
	// Load page ids with the most todos
	rows := database.NewQuery(`
		SELECT l.parentId
		FROM (
			SELECT l.parentId AS parentId,l.childAlias AS childAlias,p.todoCount AS parentTodoCount
			FROM links AS l
			JOIN pages AS p
			ON (l.parentId=p.pageId)
			WHERE p.isLiveEdit AND p.creatorId=?`, returnData.User.ID).Add(`
		) AS l
		LEFT JOIN`).AddPart(core.PageInfosTable(returnData.User)).Add(`AS pi
		ON (l.childAlias=pi.alias OR l.childAlias=pi.pageId)
		WHERE pi.seeDomainId=?`, privateDomainID).Add(`
			AND pi.type!=?`, core.CommentPageType).Add(`
		GROUP BY 1
		ORDER BY (SUM(ISNULL(pi.pageId)) + MAX(l.parentTodoCount)) DESC
		LIMIT ?`, numToLoad).ToStatement(db).Query()
	return core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
}
