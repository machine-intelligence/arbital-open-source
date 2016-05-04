// intrasitePopoverJsonHandler.go contains the handler for returning JSON with data
// to display a primary page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// intrasitePopoverJsonData contains parameters passed in via the request.
type intrasitePopoverJsonData struct {
	PageAlias string
}

var intrasitePopoverHandler = siteHandler{
	URI:         "/json/intrasitePopover/",
	HandlerFunc: intrasitePopoverJsonHandler,
}

// intrasitePopoverJsonHandler handles the request.
func intrasitePopoverJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	// Decode data
	var data intrasitePopoverJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	// Load data
	returnData := core.NewHandlerData(params.U, false)
	core.AddPageToMap(pageId, returnData.PageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}
