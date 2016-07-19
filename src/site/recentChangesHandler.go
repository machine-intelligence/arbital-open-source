// recentChangesHandler.go serves the recent changes feed.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type recentChangesData struct {
	NumToLoad     int
	CreatedBefore string
}

var recentChangesHandler = siteHandler{
	URI:         "/json/recentChanges/",
	HandlerFunc: recentChangesHandlerFunc,
	Options:     pages.PageOptions{},
}

func recentChangesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data recentChangesData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumToLoad <= 0 {
		data.NumToLoad = DefaultModeRowCount
	}

	// Load edits, edit proposals, and deleted pages
	// TODO: add newPage as a changeLog event, and then include it here.
	changeLogRows, err := loadChangeLogModeRows(db, returnData, data.NumToLoad, data.CreatedBefore,
		core.NewEditProposalChangeLog,
		core.NewEditChangeLog,
		core.DeletePageChangeLog,
		core.RevertEditChangeLog)
	if err != nil {
		return pages.Fail("Error loading changeLogRows", err)
	}

	pageToDomainSubmissionRows, err := loadPageToDomainSubmissionModeRows(db, returnData, data.NumToLoad)
	if err != nil {
		return pages.Fail("Error loading pageToDomainSubmissionRows", err)
	}

	returnData.ResultMap["modeRows"] = combineModeRows(data.NumToLoad, pageToDomainSubmissionRows, changeLogRows)

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
