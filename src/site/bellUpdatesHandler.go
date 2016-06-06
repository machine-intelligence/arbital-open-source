// bellUpdatesHandler.go serves the /notifications panel (which displays notifications, such as, 'Alexei replied to your comment').
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type bellUpdatesData struct {
	NumPagesToLoad int
}

var bellUpdatesHandler = siteHandler{
	URI:         "/json/notifications/",
	HandlerFunc: bellUpdatesHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func bellUpdatesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data bellUpdatesData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	rows, err := loadNotificationRows(db, u, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading notifications", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, rows)

	// Load and update lastBellUpdatesView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastBellUpdatesView)
	if err != nil {
		return pages.Fail("Error updating last response mode view", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	// Mark updates as seen.
	updateIds := make([]string, 0)
	for _, row := range rows {
		updateIds = append(updateIds, row.(*updateModeRow).Update.Id)
	}
	err = core.MarkUpdatesAsSeen(db, u.Id, updateIds)
	if err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}

	return pages.Success(returnData)
}
