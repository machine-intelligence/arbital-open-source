// moreRelationshipsJsonHandler.go serves JSON data to display on the additional relationships tab in the page editor.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type moreRelationshipsJsonData struct {
	PageId string
}

var moreRelationshipsHandler = siteHandler{
	URI:         "/json/moreRelationships/",
	HandlerFunc: moreRelationshipsJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// moreRelationshipsJsonHandler handles the request.
func moreRelationshipsJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data moreRelationshipsJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}

	// Load additional info for all pages
	pageOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)

	// Load recently created page ids.
	rows := database.NewQuery(`
		SELECT l.parentId
		FROM links AS l
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (pi.pageId=l.childAlias OR pi.alias=l.childAlias)
		WHERE pi.pageId=?`, data.PageId).ToStatement(db).Query()
	returnData.ResultMap["moreRelationshipIds"], err = core.LoadPageIds(rows, returnData.PageMap, pageOptions)
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
