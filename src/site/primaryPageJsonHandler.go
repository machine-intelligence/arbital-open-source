// primaryPageJsonHandler.go contains the handler for returning JSON with data
// to display a primary page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// primaryPageJsonData contains parameters passed in via the request.
type primaryPageJsonData struct {
	PageAlias string
}

var primaryPageHandler = siteHandler{
	URI:         "/json/primaryPage/",
	HandlerFunc: primaryPageJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

// primaryPageJsonHandler handles the request.
func primaryPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	returnData := core.NewHandlerData(params.U, true)

	// Decode data
	var data primaryPageJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	// Check if page is a user page
	row := db.NewStatement(`
		SELECT id
		FROM users
		WHERE id=?`).QueryRow(pageId)
	var id string
	exists, err := row.Scan(&id)
	if err != nil {
		fmt.Errorf("failed to scan for a member: %v", err)
	}
	// If page is a user page, add some values to returnData
	if exists {
		c.Infof("Page is a user page, id: %v", id)

		// Options to load the pages with
		pageOptions := (&core.PageLoadOptions{
			RedLinkCount: true,
		}).Add(core.TitlePlusLoadOptions)

		// Load recently created by me page ids.
		rows := db.NewStatement(`
			SELECT pi.pageId
			FROM pageInfos AS pi
			WHERE pi.currentEdit>0 AND pi.createdBy=? AND pi.seeGroupId=? AND pi.type!=?
			ORDER BY pi.createdAt DESC
			LIMIT ?`).Query(pageId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
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
			LIMIT ?`).Query(pageId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
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
			WHERE pi.currentEdit>0 AND p.creatorId=? AND pi.seeGroupId=? AND pi.type!=?
			GROUP BY 1
			ORDER BY MAX(p.createdAt) DESC
			LIMIT ?`).Query(pageId, params.PrivateGroupId, core.CommentPageType, indexPanelLimit)
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
			LIMIT ?`).Query(params.PrivateGroupId, pageId, core.CommentPageType, indexPanelLimit)
		returnData.ResultMap["topPagesIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
		if err != nil {
			return pages.HandlerErrorFail("error while loading recently edited by me page ids", err)
		}

		returnData.UserMap[pageId] = &core.User{Id: pageId}
	}

	// Load data
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryPageLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.ToJson())
}
