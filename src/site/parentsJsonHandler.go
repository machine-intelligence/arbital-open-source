// parentsJsonHandler.go contains the handler for returning JSON with parents pages.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"github.com/gorilla/schema"
)

// parentsJsonData contains parameters passed in via the request.
type parentsJsonData struct {
	ChildId int64
}

// parentsJsonHandler handles the request.
func parentsJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data parentsJsonData
	params.R.ParseForm()
	err := schema.NewDecoder().Decode(&data, params.R.Form)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if data.ChildId <= 0 {
		return pages.HandlerBadRequestFail("Need a valid childId", err)
	}

	// Load the parents
	pageMap := make(map[int64]*core.Page)
	userMap := make(map[int64]*core.User)
	masteryMap := make(map[int64]*core.Mastery)

	loadOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.ChildId, pageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load pages", err)
	}
	// Remove the child, since we only want to return parents.
	delete(pageMap, data.ChildId)

	returnData := createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap)
	return pages.StatusOK(returnData)
}
