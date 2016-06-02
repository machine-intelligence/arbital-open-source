// Provide data for "continue writing" mode.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type continueWritingModeData struct {
	NumPagesToLoad int
}

var continueWritingModeHandler = siteHandler{
	URI:         "/json/continueWriting/",
	HandlerFunc: continueWritingModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func continueWritingModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data continueWritingModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load my drafts
	draftRows, err := loadDraftRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load my pages tagged for edit
	taggedForEditRows, err := loadTaggedForEditRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, draftRows, taggedForEditRows)

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
