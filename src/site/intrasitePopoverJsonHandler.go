// intrasitePopoverJsonHandler.go contains the handler for returning JSON with data
// to display a primary page.
package site

import (
	"encoding/json"
	"net/http"

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
	returnData := core.NewHandlerData(u)

	// Decode data
	var data intrasitePopoverJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageID, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.Fail("Couldn't find page", err)
	}

	// Load data
	core.AddPageToMap(pageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
