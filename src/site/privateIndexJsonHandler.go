// privateIndexJsonHandler.go serves JSON data to display private index page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var privateIndexHandler = siteHandler{
	URI:         "/json/privateIndex/",
	HandlerFunc: privateIndexJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// privateIndexJsonHandler handles the request.
func privateIndexJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	privateGroupId := params.PrivateGroupId

	var err error
	returnData := newHandlerData(true)
	returnData.User = u

	// Load recently created page ids.
	rows := db.NewStatement(`
		SELECT pageId
		FROM pageInfos
		WHERE currentEdit>0 AND seeGroupId=?
		ORDER BY createdAt DESC
		LIMIT ?`).Query(privateGroupId, indexPanelLimit)
	returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.Fail("error while loading recently created page ids", err)
	}

	// Load most liked page ids.
	rows = db.NewStatement(`
		SELECT l2.pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM likes
				ORDER BY id DESC
			) AS l1
			GROUP BY userId,pageId
		) AS l2
		JOIN pageInfos AS pi
		ON (l2.pageId=pi.pageId)
		WHERE pi.seeGroupId=?
		GROUP BY l2.pageId
		ORDER BY SUM(l2.value) DESC
		LIMIT ?`).Query(privateGroupId, indexPanelLimit)
	returnData.ResultMap["mostLikedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.Fail("error while loading most liked page ids", err)
	}

	// Load recently edited page ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,max(createdAt) AS createdAt
			FROM pages
			WHERE NOT isSnapshot AND NOT isAutosave
			GROUP BY pageId
			HAVING(SUM(1) > 1)
		) AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		WHERE pi.seeGroupId=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(privateGroupId, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.Fail("error while loading recently edited page ids", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	rows = db.NewStatement(`
		SELECT pi.pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM votes
				ORDER BY id DESC
			) AS v1
			GROUP BY userId,pageId
		) AS v2
		JOIN pageInfos AS pi
		ON (v2.pageId=pi.pageId)
		WHERE pi.seeGroupId=?
		GROUP BY pi.pageId
		ORDER BY VAR_POP(v2.value) DESC
		LIMIT ?`).Query(privateGroupId, indexPanelLimit)
	returnData.ResultMap["mostControversialIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.Fail("error while loading most controversial page ids", err)
	}

	// Load pages.
	core.AddPageToMap(privateGroupId, returnData.PageMap, core.EmptyLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
