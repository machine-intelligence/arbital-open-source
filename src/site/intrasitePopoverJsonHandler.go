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

// intrasitePopoverJsonHandler handles the request.
func intrasitePopoverJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data intrasitePopoverJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	aliasToIdMap, err := core.LoadAliasToPageIdMap(db, []string{data.PageAlias})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	pageId, ok := aliasToIdMap[data.PageAlias]
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	masteryMap := make(map[int64]*core.Mastery)

	core.AddPageToMap(pageId, pageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData := createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap)
	return pages.StatusOK(returnData)
}
