// userPage.go serves the user template.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var userPageHandler = siteHandler{
	URI:         "/json/userPage/",
	HandlerFunc: userPageJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

type userPageJsonData struct {
	UserId int64 `json:",string"`
}

// userPageJsonHandler renders the user page.
func userPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	// Decode data
	var data userPageJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if data.UserId < 0 {
		return pages.HandlerBadRequestFail("Need a valid parentId", err)
	} else if data.UserId == 0 {
		data.UserId = u.Id
	}

	returnData := newHandlerData(true)
	returnData.User = u

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

	// Load recently created by me page ids.
	rows := db.NewStatement(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
	}

	// Load recently created by me comment ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=? AND pi.seeGroupId=? AND pi.type=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["recentlyCreatedCommentIds"], err =
		core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
	}

	// Load recently edited by me page ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		WHERE p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	if u.Id == data.UserId {
		pagesWithDraftIds := make([]string, 0)
		// Load pages with unpublished drafts
		rows = db.NewStatement(`
			SELECT p.pageId,p.title,p.createdAt,pi.currentEdit>0
			FROM pages AS p
			JOIN pageInfos AS pi
			ON (p.pageId = pi.pageId)
			WHERE p.creatorId=? AND pi.type!=? AND p.edit>pi.currentEdit AND pi.seeGroupId=? AND (p.text!="" OR p.title!="")
			GROUP BY p.pageId
			ORDER BY p.createdAt DESC
			LIMIT ?`).Query(data.UserId, core.CommentPageType, params.PrivateGroupId, indexPanelLimit)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var title, createdAt string
			var wasPublished bool
			err := rows.Scan(&pageId, &title, &createdAt, &wasPublished)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			pagesWithDraftIds = append(pagesWithDraftIds, fmt.Sprintf("%d", pageId))
			page := core.AddPageIdToMap(pageId, returnData.EditMap)
			if title == "" {
				title = "*Untitled*"
			}
			page.Title = title
			page.CreatedAt = createdAt
			page.WasPublished = wasPublished
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("error while loading pages with drafts ids", err)
		}
		returnData.ResultMap["pagesWithDraftIds"] = pagesWithDraftIds

		// Load page ids with the most todos.
		rows = db.NewStatement(`
			SELECT l.parentId
			FROM (
				SELECT l.parentId AS parentId,l.childAlias AS childAlias,p.todoCount AS parentTodoCount
				FROM links AS l
				JOIN pages AS p
				ON (l.parentId=p.pageId)
				WHERE p.creatorId=? AND p.isCurrentEdit
			) AS l
			LEFT JOIN pageInfos AS pi
			ON (l.childAlias=pi.alias OR l.childAlias=pi.pageId)
			WHERE pi.seeGroupId=? AND pi.type!=?
			GROUP BY 1
			ORDER BY (SUM(ISNULL(pi.pageId)) + MAX(l.parentTodoCount)) DESC
			LIMIT ?`).Query(data.UserId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
		returnData.ResultMap["mostTodosIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.HandlerErrorFail("error while loading most todos page ids", err)
		}
	}

	// Load top pages by me
	rows = db.NewStatement(`
		SELECT pi.pageId
		FROM pageInfos AS pi
		JOIN (
			/* Get the last like per page for each user */
			SELECT pageId,userId,value
			FROM likes
			GROUP BY 1,2
		) AS l
		ON (pi.pageId=l.pageId)
		WHERE pi.seeGroupId=? AND pi.editGroupId=? AND pi.type!=?
		GROUP BY 1
		ORDER BY SUM(l.value) DESC
		LIMIT ?`).Query(params.PrivateGroupId, data.UserId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["topPagesIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited by me page ids", err)
	}

	// Load pages.
	returnData.UserMap[data.UserId] = &core.User{Id: data.UserId}
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
