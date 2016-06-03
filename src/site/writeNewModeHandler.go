// Provide data for "write new" mode.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type writeNewModeData struct {
	NumPagesToLoad int
}

var writeNewModeHandler = siteHandler{
	URI:         "/json/writeNew/",
	HandlerFunc: writeNewModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func writeNewModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data writeNewModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load redlinks in math
	redLinks, err := loadRedLinkRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, redLinks)

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
