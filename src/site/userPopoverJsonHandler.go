// userPopoverJsonHandler.go contains the handler for returning JSON with user data
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// userPopoverJsonData contains parameters passed in via the request.
type userPopoverJsonData struct {
	UserId int64 `json:",string"`
}

var userPopoverHandler = siteHandler{
	URI:         "/json/userPopover/",
	HandlerFunc: userPopoverJsonHandler,
}

// userPopoverJsonHandler handles the request.
func userPopoverJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data userPopoverJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Load data
	returnData := newHandlerData(false)
	core.AddUserToMap(data.UserId, returnData.UserMap)
	// This page is the user page, this will load user summary
	core.AddPageToMap(data.UserId, returnData.PageMap, core.IntrasitePopoverLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		pages.HandlerErrorFail("Pipeline error", err)
	}

	db.C.Debugf("in userPopoverJsonHandler")
	db.C.Debugf("data.UserId: %v", data.UserId)
	db.C.Debugf("returnData.UserMap: %v", returnData.UserMap)
	db.C.Debugf("returnData.UserMap[data.UserId]: %v", returnData.UserMap[data.UserId])

	return pages.StatusOK(returnData.toJson())
}
