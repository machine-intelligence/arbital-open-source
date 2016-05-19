// discussionModeHandler.go serves the /discussion panel.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type discussionModeData struct {
	NumPagesToLoad int
}

var discussionModeHandler = siteHandler{
	URI:         "/json/discussionMode/",
	HandlerFunc: discussionModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func discussionModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data discussionModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load all comments of interest
	commentRows, err := loadCommentModeRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("failed to load comment ids", err)
	}

	// Load all marks of interest
	markRows, err := loadMarkModeRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("failed to load mark ids", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, commentRows, markRows)

	// Load and update LastDiscussionView for this user
	returnData.ResultMap["lastView"], err = LoadAndUpdateLastView(db, u, LastDiscussionModeView)
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
