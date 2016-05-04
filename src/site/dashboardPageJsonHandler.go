// dashboardPage.go serves the dashboard template.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var dashboardPageHandler = siteHandler{
	URI:         "/json/dashboardPage/",
	HandlerFunc: dashboardPageJsonHandler,
	Options: pages.PageOptions{
		RequireLogin:    true,
		LoadUpdateCount: true,
	},
}

type dashboardPageJsonData struct {
}

// dashboardPageJsonHandler renders the dashboard page.
func dashboardPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data dashboardPageJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

	// Load recently created by me comment ids
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=?`, u.Id).Add(`
			AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
		ORDER BY pi.createdAt DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["recentlyCreatedCommentIds"], err =
		core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
	}

	// Load recently created and edited by me page ids
	rows = database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId)
		WHERE p.creatorId=?`, u.Id).Add(`
			AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
			AND pi.type!=?`, core.CommentPageType).Add(`
		GROUP BY 1
		ORDER BY MAX(p.createdAt) DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	pagesWithDraftIds := make([]string, 0)
	// Load pages with unpublished drafts
	rows = database.NewQuery(`
			SELECT p.pageId,p.title,p.createdAt,pi.currentEdit>0,pi.isDeleted
			FROM pages AS p
			JOIN`).AddPart(core.PageInfosTableAll(u)).Add(`AS pi
			ON (p.pageId = pi.pageId)
			WHERE p.creatorId=?`, u.Id).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
				AND pi.seeGroupId=?`, params.PrivateGroupId).Add(`
				AND p.edit>pi.currentEdit AND (p.text!="" OR p.title!="")
			GROUP BY p.pageId
			ORDER BY p.createdAt DESC
			LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		var title, createdAt string
		var wasPublished bool
		var isDeleted bool
		err := rows.Scan(&pageId, &title, &createdAt, &wasPublished, &isDeleted)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		core.AddPageToMap(pageId, returnData.PageMap, pageOptions)
		pagesWithDraftIds = append(pagesWithDraftIds, pageId)
		page := core.AddPageIdToMap(pageId, returnData.EditMap)
		if title == "" {
			title = "*Untitled*"
		}
		page.Title = title
		page.CreatedAt = createdAt
		page.WasPublished = wasPublished
		page.IsDeleted = isDeleted
		return nil
	})
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages with drafts ids", err)
	}
	returnData.ResultMap["pagesWithDraftIds"] = pagesWithDraftIds

	// Load page ids with the most todos
	rows = database.NewQuery(`
		SELECT l.parentId
		FROM (
			SELECT l.parentId AS parentId,l.childAlias AS childAlias,p.todoCount AS parentTodoCount
			FROM links AS l
			JOIN pages AS p
			ON (l.parentId=p.pageId)
			WHERE p.isLiveEdit AND p.creatorId=?`, u.Id).Add(`
		) AS l
		LEFT JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (l.childAlias=pi.alias OR l.childAlias=pi.pageId)
		WHERE pi.seeGroupId=?`, params.PrivateGroupId).Add(`
			AND pi.type!=?`, core.CommentPageType).Add(`
		GROUP BY 1
		ORDER BY (SUM(ISNULL(pi.pageId)) + MAX(l.parentTodoCount)) DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["mostTodosIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading most todos page ids", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}
