// userPopoverJsonHandler.go contains the handler for returning JSON with user data
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// userPopoverJsonData contains parameters passed in via the request.
type userPopoverJsonData struct {
	UserId string
}

var userPopoverHandler = siteHandler{
	URI:         "/json/userPopover/",
	HandlerFunc: userPopoverJsonHandler,
}

// userPopoverJsonHandler handles the request.
func userPopoverJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U)

	// Decode data
	var data userPopoverJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Load data
	core.AddUserToMap(data.UserId, returnData.UserMap)
	// This page is the user page, this will load user summary
	core.AddPageToMap(data.UserId, returnData.PageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
