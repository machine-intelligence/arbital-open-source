// childrenJsonHandler.go contains the handler for returning JSON with children pages.
package site

import (
	"fmt"

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
	pageMap[data.ParentId] = &core.Page{PageId: data.ParentId}
	err = loadChildrenIds(db, pageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load children", err)
	}
	// Remove parent, since we only want to return children.
	delete(pageMap, data.ParentId)

	// Load pages.
	err = core.LoadPages(db, pageMap, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Load likes.
	err = loadAuxPageData(db, u.Id, pageMap, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load aux data", err)
	}

	// Return the page in JSON format.
	strPageMap := make(map[string]*core.Page)
	for k, v := range pageMap {
		strPageMap[fmt.Sprintf("%d", k)] = v
	}
	return pages.StatusOK(strPageMap)
}
