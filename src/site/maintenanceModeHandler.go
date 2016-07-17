// maintenanceModeHandler.go serves the /maintain panel (which displays maintenance updates, such as, 'Alexei edited your page').
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type maintenanceModeData struct {
	NumPagesToLoad int
}

var maintenanceModeHandler = siteHandler{
	URI:         "/json/maintain/",
	HandlerFunc: maintenanceModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func maintenanceModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data maintenanceModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	rows, err := loadMaintenanceUpdateRows(db, u, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading maintenance updates", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumPagesToLoad, rows)

	// Load and update lastMaintenanceModeView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastMaintenanceModeView)
	if err != nil {
		return pages.Fail("Error updating last maintenance mode view", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	// Mark updates as seen.
	updateIds := make([]string, 0)
	for _, row := range rows {
		updateIds = append(updateIds, row.(*updateModeRow).Update.ID)
	}
	err = core.MarkUpdatesAsSeen(db, u.ID, core.GetMaintenanceUpdateTypes())
	if err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}

	return pages.Success(returnData)
}
