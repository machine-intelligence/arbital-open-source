// lensJsonHandler.go contains the handler for returning JSON with data to display a lens.
package site

import (
	"encoding/json"
	"net/http"

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
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data lensJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.Fail("Couldn't find page", err)
	}

	// Load data
	core.AddPageToMap(pageId, returnData.PageMap, core.LensFullLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
