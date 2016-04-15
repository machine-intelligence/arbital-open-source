// parentsJsonHandler.go contains the handler for returning JSON with parents pages.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// parentsJsonData contains parameters passed in via the request.
type parentsJsonData struct {
	ChildId string
}

var parentsHandler = siteHandler{
	URI:         "/json/parents/",
	HandlerFunc: parentsJsonHandler,
}

// parentsJsonHandler handles the request.
func parentsJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Decode data
	var data parentsJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if !core.IsIdValid(data.ChildId) {
		return pages.HandlerBadRequestFail("Need a valid childId", err)
	}

	// Load the parents
	returnData := core.NewHandlerData(params.U, false)

	loadOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.ChildId, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load pages", err)
	}
	// Remove the child, since we only want to return parents.
	delete(returnData.PageMap, data.ChildId)

	return pages.StatusOK(returnData)
}
