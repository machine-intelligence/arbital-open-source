// titleJsonHandler.go contains the handler for returning JSON with data to display a title.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// titleJsonData contains parameters passed in via the request.
type titleJsonData struct {
	PageAlias string
}

var titleHandler = siteHandler{
	URI:         "/json/title/",
	HandlerFunc: titleJsonHandler,
}

// titleJsonHandler handles the request.
func titleJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	db.C.Debugf("in titleJsonHandler")

	// Decode data
	var data titleJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	db.C.Debugf("data.PageAlias: %v", data.PageAlias)

	// Get actual page id
	returnData := newHandlerData(false)
	pageId, ok, err := core.LoadAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		db.C.Debugf("failed to load alias")
		// Don't fail because sometimes the editor calls this with bad aliases, but
		// we don't want to generated messages on the FE
		return pages.StatusOK(returnData.toJson())
	}

	db.C.Debugf("successfully loaded alias")
	db.C.Debugf("pageId: %v", pageId)

	// Load data
	core.AddPageIdToMap(pageId, returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
