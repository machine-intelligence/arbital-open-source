// updatesPage.go serves the update page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	commonPageData
}

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

	// Load the updates and populate page & user maps
	returnData := newHandlerData()
	updateRows, err := core.LoadUpdateRows(db, u.Id, returnData.PageMap, returnData.UserMap, false)
	if err != nil {
		return pages.Fail("failed to load updates", err)
	}

	// Load data
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	// Now that we have loaded last visit time for all pages,
	// go through all the update rows and group them.
	returnData.ResultMap["updateGroups"] = core.ConvertUpdateRowsToGroups(updateRows, returnData.PageMap)

	// Load subscriptions to users
	err = core.LoadUserSubscriptions(db, u.Id, returnData.UserMap)
	if err != nil {
		return pages.Fail("error while loading subscriptions to users", err)
	}

	// Zero out all counts.
	statement := db.NewStatement(`
		UPDATE updates
		SET newCount=0
		WHERE userId=?`)
	if _, err = statement.Exec(u.Id); err != nil {
		return pages.Fail("Couldn't mark updates seen", err)
	}
	return pages.StatusOK(returnData.toJson())
}
