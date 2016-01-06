// userPage.go serves the user template.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
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
		WHERE pi.currentEdit>0 AND p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
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
		WHERE pi.currentEdit>0 AND p.creatorId=? AND pi.seeGroupId=? AND pi.type=?
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
		WHERE pi.currentEdit>0 AND p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.UserId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	// Load top pages by me
	rows = db.NewStatement(`
		SELECT pi.pageId
		FROM pageInfos AS pi
		JOIN (
			SELECT *
			FROM (
				SELECT *
				FROM likes
				ORDER BY id DESC
			) AS l1
			GROUP BY userId,pageId
		) AS l2
		ON (pi.pageId=l2.pageId)
		WHERE pi.currentEdit>0 AND pi.seeGroupId=? AND pi.editGroupId=? AND pi.type!=?
		GROUP BY 1
		ORDER BY SUM(l2.value) DESC
		LIMIT ?`).Query(params.PrivateGroupId, data.UserId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["topPagesIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited by me page ids", err)
	}

	// Load pages.
	core.AddPageToMap(data.UserId, returnData.PageMap, core.PrimaryPageLoadOptions)
	returnData.UserMap[data.UserId] = &core.User{Id: data.UserId}
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
