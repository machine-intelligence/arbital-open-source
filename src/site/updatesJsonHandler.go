// updatesPage.go serves the update page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var updatesHandler = siteHandler{
	URI:         "/json/updates/",
	HandlerFunc: updatesJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updatesJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Load the updates and populate page & user maps
	updateRows, err := core.LoadUpdateRows(db, u, returnData, false)
	if err != nil {
		return pages.Fail("failed to load updates", err)
	}

	// Load data
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	// Now that we have loaded last visit time for all pages,
	// go through all the update rows and group them.
	returnData.ResultMap["updateGroups"] = core.ConvertUpdateRowsToGroups(updateRows, returnData.PageMap)

	// Zero out all counts.
	statement := db.NewStatement(`
		UPDATE updates
		SET unseen=FALSE
		WHERE userId=?`)
	if _, err = statement.Exec(u.Id); err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}
	return pages.Success(returnData)
}
