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

	// Load the parents.
	pageMap := make(map[int64]*core.Page)
	pageMap[data.ChildId] = &core.Page{PageId: data.ChildId}
	err = core.LoadParentsIds(db, pageMap, &core.LoadParentsIdsOptions{LoadHasParents: true})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load parent ids", err)
	}
	// Remove child, since we only want to return parents.
	delete(pageMap, data.ChildId)

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Load auxillary data.
	err = core.LoadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't retrieve page likes", err)
	}

	returnData := createReturnData(pageMap)
	return pages.StatusOK(returnData)
}
