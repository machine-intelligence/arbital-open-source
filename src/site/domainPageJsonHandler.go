// domainPageJsonHandler.go serves JSON data to display domain index page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	indexPanelLimit = 10
)

type domainPageJsonData struct {
	DomainAlias string
}

var domainPageHandler = siteHandler{
	URI:         "/json/domainPage/",
	HandlerFunc: domainPageJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

// domainPageJsonHandler handles the request.
func domainPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data domainPageJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	returnData := core.NewHandlerData(params.U, true)

	// Get constraint
	var constraintPart *database.QueryPart
	if data.DomainAlias != "" {
		domainId, ok, err := core.LoadAliasToPageId(db, u, data.DomainAlias)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't convert alias", err)
		}
		if !ok {
			return pages.HandlerErrorFail(fmt.Sprintf("Couldn't find the domain: %s", data.DomainAlias), nil)
		}
		core.AddPageToMap(domainId, returnData.PageMap, core.PrimaryPageLoadOptions)
		constraintPart = database.NewQuery("AND pd.domainId=?", domainId)
		returnData.ResultMap["domainId"] = domainId
	} else {
		if !core.IsIdValid(params.PrivateGroupId) {
			return pages.HandlerBadRequestFail("Need domain alias or need to be in a private domain", err)
		}
		core.AddPageToMap(params.PrivateGroupId, returnData.PageMap, core.PrimaryPageLoadOptions)
		constraintPart = database.NewQuery("AND pi.seeGroupId=?", params.PrivateGroupId)
		returnData.ResultMap["domainId"] = params.PrivateGroupId
	}

	// Load additional info for all pages
	pageOptions := (&core.PageLoadOptions{
		SubpageCounts: true,
	}).Add(core.TitlePlusLoadOptions)

	// Load recently created page ids.
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId)
		LEFT JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		WHERE p.isLiveEdit AND pi.type!=?`, core.CommentPageType).AddPart(constraintPart).Add(`
		ORDER BY pi.createdAt DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["recentlyCreatedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently created page ids", err)
	}

	// Load most liked page ids.
	rows = database.NewQuery(`
		SELECT pi.pageId
		FROM likes AS l2
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (l2.likeableId=pi.likeableId)
		LEFT JOIN pageDomainPairs AS pd
		ON (pi.pageId=pd.pageId)
		WHERE TRUE`).AddPart(constraintPart).Add(`
		GROUP BY l2.likeableId
		ORDER BY SUM(value) DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["mostLikedIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading most liked page ids", err)
	}

	// Load hot page ids (recent comments + edits).
	rows = database.NewQuery(`
		SELECT p.pageId
		FROM (
			/* Count recent edits */
			SELECT pageId,sum(1) AS tally
			FROM pages
			WHERE NOT isSnapshot AND NOT isAutosave AND DATEDIFF(now(),createdAt) <= 7
			GROUP BY 1
			UNION ALL
			/* Count recent new comments */
			SELECT pp.parentId,sum(1) AS tally
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			JOIN pagePairs AS pp
			ON (pi.pageId=pp.childId)
			WHERE DATEDIFF(now(),pi.createdAt) <= 7 AND pp.type=?`, core.ParentPagePairType).Add(`
				AND pi.type=?`, core.CommentPageType).Add(`
			GROUP by 1
		) AS p
		LEFT JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.pageId)
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (p.pageId=pi.pageId)
		WHERE pi.type!=?`, core.CommentPageType).AddPart(constraintPart).Add(`
		GROUP BY 1
		ORDER BY SUM(tally) DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["hotIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading recently edited page ids", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	rows = database.NewQuery(`
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
		LEFT JOIN pageDomainPairs AS pd
		ON (v2.pageId=pd.pageId)
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (pd.pageId=pi.pageId)
		WHERE pi.hasVote`).AddPart(constraintPart).Add(`
		GROUP BY pd.pageId
		HAVING COUNT(v2.value) > 1
		ORDER BY VAR_POP(v2.value) DESC
		LIMIT ?`, indexPanelLimit).ToStatement(db).Query()
	returnData.ResultMap["mostControversialIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("error while loading most controversial page ids", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}
