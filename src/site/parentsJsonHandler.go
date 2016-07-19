// parentsJsonHandler.go contains the handler for returning JSON with parents pages.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// parentsJsonData contains parameters passed in via the request.
type parentsJSONData struct {
	ChildID string
}

var parentsHandler = siteHandler{
	URI:         "/json/parents/",
	HandlerFunc: parentsJSONHandler,
}

// parentsJsonHandler handles the request.
func parentsJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U)

	// Decode data
	var data parentsJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.ChildID) {
		return pages.Fail("Need a valid childId", err).Status(http.StatusBadRequest)
	}

	// Load the parents
	loadOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.ChildID, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Couldn't load pages", err)
	}
	// Remove the child, since we only want to return parents.
	delete(returnData.PageMap, data.ChildID)

	return pages.Success(returnData)
}
