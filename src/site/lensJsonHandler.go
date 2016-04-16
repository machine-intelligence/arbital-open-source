// lensJsonHandler.go contains the handler for returning JSON with data to display a lens.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// lensJsonData contains parameters passed in via the request.
type lensJsonData struct {
	PageAlias string
}

var lensHandler = siteHandler{
	URI:         "/json/lens/",
	HandlerFunc: lensJsonHandler,
}

// lensJsonHandler handles the request.
func lensJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Decode data
	var data lensJsonData
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

	// Load data
	returnData := core.NewHandlerData(params.U, false)
	core.AddPageToMap(pageId, returnData.PageMap, core.LensFullLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}
