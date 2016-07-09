// recentChangesHandler.go serves the recent changes feed.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type recentRelationshipChangesData struct {
	NumToLoad     int
	CreatedBefore string
}

var recentRelationshipChangesHandler = siteHandler{
	URI:         "/json/recentRelationshipChanges/",
	HandlerFunc: recentRelationshipChangesHandlerFunc,
	Options:     pages.PageOptions{},
}

func recentRelationshipChangesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data recentRelationshipChangesData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumToLoad <= 0 {
		data.NumToLoad = DefaultModeRowCount
	}
	if data.CreatedBefore == "" {
		data.CreatedBefore = database.Now()
	}

	changeLogRows, err := loadChangeLogModeRows(db, returnData, data.NumToLoad, data.CreatedBefore,
		core.NewParentChangeLog,
		core.DeleteParentChangeLog,
		core.NewTagChangeLog,
		core.DeleteTagChangeLog,
		core.NewRequirementChangeLog,
		core.DeleteRequirementChangeLog)
	if err != nil {
		return pages.Fail("Error loading changeLogRows", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumToLoad, changeLogRows)

	// Load and update LastReadModeView for this user
	returnData.ResultMap["lastView"], err = core.LoadAndUpdateLastView(db, u, core.LastRecentChangesView)
	if err != nil {
		return pages.Fail("Error updating last recent changes view", err)
	}

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
