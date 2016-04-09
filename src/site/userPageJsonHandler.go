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
	UserAlias string
}

// userPageJsonHandler renders the user page.
func userPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U, true)

	// Decode data
	var data userPageJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if data.UserAlias == "" {
		return pages.HandlerBadRequestFail("Need a user alias", nil)
	}

	// Get actual user id
	userId, ok, err := core.LoadAliasToPageId(db, data.UserAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find user", err)
	}

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

	// Load recently created by me page ids.
	rows := db.NewStatement(`
		SELECT pi.pageId
		FROM pageInfos AS pi
		WHERE pi.currentEdit>0 AND NOT pi.isDeleted AND pi.createdBy=? AND pi.seeGroupId=? AND pi.type!=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(userId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
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
		WHERE pi.currentEdit>0 AND NOT pi.isDeleted AND p.creatorId=? AND pi.seeGroupId=? AND pi.type=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(userId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
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
		ON (p.pageId=pi.pageId)
		WHERE pi.currentEdit>0 AND NOT pi.isDeleted AND p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
		GROUP BY 1
		ORDER BY MAX(p.createdAt) DESC
		LIMIT ?`).Query(userId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	// Load top pages by me
	rows = db.NewStatement(`
		SELECT pi.pageId
		FROM pageInfos AS pi
		JOIN likes l2
		ON (pi.pageId=l2.pageId)
		WHERE pi.currentEdit>0 AND NOT pi.isDeleted AND pi.seeGroupId=? AND pi.editGroupId=? AND pi.type!=?
		GROUP BY 1
		ORDER BY SUM(l2.value) DESC
		LIMIT ?`).Query(params.PrivateGroupId, userId, core.CommentPageType, indexPanelLimit)
	returnData.ResultMap["topPagesIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited by me page ids", err)
	}

	// Load pages.
	core.AddPageToMap(userId, returnData.PageMap, core.PrimaryPageLoadOptions)
	returnData.UserMap[userId] = &core.User{Id: userId}
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.ToJson())
}
