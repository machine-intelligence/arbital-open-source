// moreRelationshipsJsonHandler.go serves JSON data to display on the additional relationships tab in the page editor.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type moreRelationshipsJSONData struct {
	PageAlias            string
	RestrictToMathDomain bool
}

var moreRelationshipsHandler = siteHandler{
	URI:         "/json/moreRelationships/",
	HandlerFunc: moreRelationshipsJSONHandler,
	Options:     pages.PageOptions{},
}

// moreRelationshipsJsonHandler handles the request.
func moreRelationshipsJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data moreRelationshipsJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsAliasValid(data.PageAlias) {
		return pages.Fail("Invalid page id or alias", nil).Status(http.StatusBadRequest)
	}

	// Load pages that link to this page.
	query := database.NewQuery(`
		SELECT l.parentId
		FROM links AS l
		JOIN `).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
		ON (l.parentId=pi.pageId OR l.parentId=pi.alias)`)
	if data.RestrictToMathDomain {
		query.Add(`
			JOIN pageDomainPairs as pdp
			ON pdp.pageId=l.parentId AND pdp.domainId=?`, core.MathDomainID)
	}
	query.Add(`WHERE l.childAlias=?`, data.PageAlias)

	rows := query.ToStatement(db).Query()
	loadOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)
	returnData.ResultMap["moreRelationshipIds"], err = core.LoadPageIds(rows, returnData.PageMap, loadOptions)
	if err != nil {
		return pages.Fail("error while loading links", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
