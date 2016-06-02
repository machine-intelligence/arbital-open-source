// hedonsModeHandler.go serves the /achievements panel (which displays hedons updates, such as, 'Alexei liked your page').
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type hedonsModeData struct {
	NumPagesToLoad int
}

var hedonsModeHandler = siteHandler{
	URI:         "/json/hedons/",
	HandlerFunc: hedonsModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func hedonsModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data hedonsModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load new likes on my pages and comments
	likesRows, err := loadLikesModeRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading new likes", err)
	}

	// Load new likes on my edits
	changeLikesRows, err := loadChangeLikesModeRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading change likes", err)
	}

	// Load requisites taught
	reqsTaughtRows, err := loadReqsTaughtModeRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading requisites taught", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, likesRows, changeLikesRows, reqsTaughtRows)

	// Load and update lastAchievementsView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastAchievementsModeView)
	if err != nil {
		return pages.Fail("Error updating last achievements view", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
