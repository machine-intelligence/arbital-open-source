// readModeHandler.go serves the /read panel.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

const (
	FeaturedReadModeType = "featured"
	NewReadModeType      = "new"
)

type readModeData struct {
	NumPagesToLoad int
	Type           string
}

var readModeHandler = siteHandler{
	URI:         "/json/readMode/",
	HandlerFunc: readModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func readModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data readModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load the page ids
	var pageIds ModeRows
	if data.Type == FeaturedReadModeType {
		pageIds, err = loadFeaturedPagesModeRows(db, returnData, data.NumPagesToLoad)
	} else if data.Type == NewReadModeType {
		pageIds, err = loadNewPagesModeRows(db, returnData, data.NumPagesToLoad)
	}
	if err != nil {
		return pages.Fail("failed to load page ids", err)
	}
	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, pageIds)

	// Load and update LastReadModeView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastReadModeView)
	if err != nil {
		return pages.Fail("Error updating last read mode view", err)
	}

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
