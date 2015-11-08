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
		FROM (
			SELECT pageId,edit,seeGroupId
			FROM pages
			WHERE creatorId=? AND type!="comment"
		) AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit AND p.seeGroupId=?)
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, indexPanelLimit)
	returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
	}

	// Load recently edited by me page ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,max(edit) AS maxEdit,min(edit) AS minEdit,
				max(createdAt) AS createdAt,max(seeGroupId) AS seeGroupId
			FROM pages
			WHERE creatorId=? AND type!="comment" AND NOT isAutosave
			GROUP BY 1
		) AS p
		WHERE maxEdit>minEdit AND p.seeGroupId=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	if u.Id == data.UserId {
		pagesWithDraftIds := make([]string, 0)
		// Load pages with unpublished drafts
		rows = db.NewStatement(`
			SELECT p.pageId,p.title,p.createdAt,i.currentEdit>0
			FROM pages AS p
			JOIN pageInfos AS i
			ON (p.pageId = i.pageId)
			WHERE p.creatorId=? AND p.type!=? AND p.edit>i.currentEdit AND p.seeGroupId=?
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
				WHERE p.creatorId=? AND p.type!="comment" AND p.isCurrentEdit AND p.seeGroupId=?
			) AS l
			LEFT JOIN pages AS p
			ON (l.childAlias=p.alias OR l.childAlias=p.pageId)
			GROUP BY 1
			ORDER BY (SUM(ISNULL(p.pageId)) + max(l.parentTodoCount)) DESC
			LIMIT ?`).Query(data.UserId, params.PrivateGroupId, indexPanelLimit)
		returnData.ResultMap["mostTodosIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
		if err != nil {
			return pages.HandlerErrorFail("error while loading most todos page ids", err)
		}
	}

	// Load recently edited by me comment ids
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE creatorId=? AND type="comment" AND seeGroupId=? AND isCurrentEdit
			GROUP BY pageId
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, indexPanelLimit)
	returnData.ResultMap["recentlyEditedCommentIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
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
