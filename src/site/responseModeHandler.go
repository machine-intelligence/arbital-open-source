// responseModeHandler.go serves the /notifications panel (which displays notifications, such as, 'Alexei replied to your comment').
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type responseModeData struct {
	NumPagesToLoad int
}

var responseModeHandler = siteHandler{
	URI:         "/json/notifications/",
	HandlerFunc: responseModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func responseModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data responseModeData
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

	// Load and update lastResponseModeView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastResponseModeView)
	if err != nil {
		return pages.Fail("Error updating last response mode view", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	updateIds := make([]string, 0)
	for _, row := range rows {
		updateIds = append(updateIds, row.(*notificationRow).Update.Id)
	}

	// Mark updates as seen.
	statement := database.NewQuery(`
		UPDATE updates
		SET seen=TRUE
		WHERE userId=?`, u.Id).Add(`
			AND id IN`).AddArgsGroupStr(updateIds).ToStatement(db)
	if _, err = statement.Exec(); err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}
	return pages.Success(returnData)
}
