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
	userMap := make(map[int64]*core.User)
	masteryMap := make(map[int64]*core.Mastery)

	loadOptions := (&core.PageLoadOptions{
		Children:                true,
		HasGrandChildren:        true,
		RedLinkCountForChildren: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.ParentId, pageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline failed", err)
	}
	// Remove parent, since we only want to return children.
	delete(pageMap, data.ParentId)

	returnData := createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap)
	return pages.StatusOK(returnData)
}
