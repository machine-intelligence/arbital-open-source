// Load and return list of unassessed pages

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type unassessedPagesData struct {
}

var unassessedPagesHandler = siteHandler{
	URI:         "/json/unassessedPages/",
	HandlerFunc: unassessedPagesHandlerFunc,
	Options:     pages.PageOptions{},
}

func unassessedPagesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data unassessedPagesData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	numPagesToLoad := DefaultModeRowCount
	pageOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)

	// Load page ids
	rows := database.NewQuery(`
		SELECT t.pageId
		FROM (
			SELECT pi.pageId AS pageId
			FROM pageInfos AS pi
			LEFT JOIN pagePairs AS pp
			ON (pi.pageId=pp.childId)
			WHERE pi.editDomainId=?`, core.MathDomainID).Add(`
				/* Check that this page doesn't have a quality tag */
				AND pp.type=?`, core.TagPagePairType).Add(`
				AND`).AddPart(core.PageInfosFilter(u)).Add(`
			GROUP BY 1
			HAVING SUM(pp.parentId IN (
					SELECT pp2.childId
					FROM pagePairs AS pp2
					WHERE pp2.type=? AND pp2.parentId=?`, core.ParentPagePairType, core.QualityMetaTagsPageID).Add(`
				)) <= 0
		) AS t
		GROUP BY 1
		ORDER BY SUM(1) DESC
		LIMIT ?`, numPagesToLoad).ToStatement(db).Query()
	returnData.ResultMap["pageIds"], err = core.LoadPageIDs(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.Fail("Error loading pageIds", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
