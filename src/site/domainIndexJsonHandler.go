// domainIndexJsonHandler.go serves JSON data to display domain index page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

const (
	indexPanelLimit = 10
)

type domainIndexJsonData struct {
	DomainAlias string
}

var domainIndexHandler = siteHandler{
	URI:         "/json/domainIndex/",
	HandlerFunc: domainIndexJsonHandler,
}

// domainIndexJsonHandler handles the request.
func domainIndexJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := newHandlerData()

	// Decode data
	var data domainIndexJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual domain id
	domainAlias := data.DomainAlias
	domainId, ok, err := core.LoadAliasToPageId(db, domainAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail(fmt.Sprintf("Couldn't find the domain: %s", domainAlias), nil)
	}

	// Load recently created page ids.
	rows := db.NewStatement(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE p.isCurrentEdit AND pd.domainId=?
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(domainId, indexPanelLimit)
	returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
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
		JOIN pageDomainPairs AS pd
		ON (l2.pageId=pd.pageId)
		WHERE pd.domainId=?
		GROUP BY l2.pageId
		ORDER BY SUM(value) DESC
		LIMIT ?`).Query(domainId, indexPanelLimit)
	returnData.ResultMap["mostLikedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading most liked page ids", err)
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
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE pd.domainId=?
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(domainId, indexPanelLimit)
	returnData.ResultMap["recentlyEditedIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	rows = db.NewStatement(`
		SELECT pd.pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM votes
				ORDER BY id DESC
			) AS v1
			GROUP BY userId,pageId
		) AS v2
		JOIN pageDomainPairs AS pd
		ON (v2.pageId=pd.pageId)
		WHERE pd.domainId=?
		GROUP BY pd.pageId
		ORDER BY VAR_POP(v2.value) DESC
		LIMIT ?`).Query(domainId, indexPanelLimit)
	returnData.ResultMap["mostControversialIds"], err = core.LoadPageIds(rows, returnData.PageMap, core.TitlePlusLoadOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading most controversial page ids", err)
	}

	// Load pages.
	core.AddPageToMap(domainId, returnData.PageMap, core.EmptyLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
