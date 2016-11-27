// titleJsonHandler.go contains the handler for returning JSON with data to display a title.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// titleJsonData contains parameters passed in via the request.
type titleJSONData struct {
	PageAlias string
}

var titleHandler = siteHandler{
	URI:         "/json/title/",
	HandlerFunc: titleJSONHandler,
}

// titleJsonHandler handles the request.
func titleJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data titleJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageID, ok, err := core.LoadAliasToPageID(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	}
	if !ok {
		// Don't fail because sometimes the editor calls this with bad aliases, but
		// we don't want to generated messages on the FE
		return pages.Success(returnData)
	}

	// Load data
	core.AddPageIDToMap(pageID, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
