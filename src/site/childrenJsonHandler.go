// childrenJsonHandler.go contains the handler for returning JSON with children pages.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"github.com/gorilla/schema"
)

// childrenJsonData contains parameters passed in to create a page.
type childrenJsonData struct {
	ParentId int64
}

// childrenJsonHandler handles requests to create a new page.
func childrenJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data childrenJsonData
	params.R.ParseForm()
	err := schema.NewDecoder().Decode(&data, params.R.Form)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if data.ParentId <= 0 {
		return pages.HandlerBadRequestFail("Need a valid parentId", err)
	}

	// Load the children.
	pageMap := make(map[int64]*core.Page)
	core.AddPageIdToMap(data.ParentId, pageMap)
	err = core.LoadChildrenIds(db, pageMap, &core.LoadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load children", err)
	}
	// Remove parent, since we only want to return children.
	delete(pageMap, data.ParentId)

	// Load pages.
	err = core.LoadPages(db, pageMap, u, nil)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Load likes.
	err = core.LoadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load aux data", err)
	}

	returnData := createReturnData(pageMap)
	return pages.StatusOK(returnData)
}
