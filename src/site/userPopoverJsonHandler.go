// userPopoverJsonHandler.go contains the handler for returning JSON with user data
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// userPopoverJsonData contains parameters passed in via the request.
type userPopoverJSONData struct {
	UserID string
}

var userPopoverHandler = siteHandler{
	URI:         "/json/userPopover/",
	HandlerFunc: userPopoverJSONHandler,
}

// userPopoverJsonHandler handles the request.
func userPopoverJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U)

	// Decode data
	var data userPopoverJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load data
	core.AddUserToMap(data.UserID, returnData.UserMap)
	// This page is the user page, this will load user summary
	core.AddPageToMap(data.UserID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
