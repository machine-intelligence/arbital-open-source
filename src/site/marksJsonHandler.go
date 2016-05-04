// marksJsonHandler.go returns marks for a given page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// marksJsonData contains parameters passed in via the request.
type marksJsonData struct {
	PageId string
}

var marksHandler = siteHandler{
	URI:         "/json/marks/",
	HandlerFunc: marksJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// marksJsonHandler handles the request.
func marksJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U)

	// Decode data
	var data marksJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Need a valid pageId", err)
	}

	// Load the marks
	loadOptions := &core.PageLoadOptions{
		AllMarks: true,
	}
	core.AddPageToMap(data.PageId, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load pages", err)
	}
	return pages.StatusOK(returnData)
}
