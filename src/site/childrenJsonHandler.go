// childrenJsonHandler.go contains the handler for returning JSON with children pages.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// childrenJsonData contains parameters passed in to create a page.
type childrenJSONData struct {
	ParentID string
}

var childrenHandler = siteHandler{
	URI:         "/json/children/",
	HandlerFunc: childrenJSONHandler,
}

// childrenJsonHandler handles requests to create a new page.
func childrenJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Decode data
	var data childrenJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.ParentID) {
		return pages.Fail("Need a valid parentId", nil).Status(http.StatusBadRequest)
	}

	returnData := core.NewHandlerData(params.U)

	// Load the children.
	loadOptions := (&core.PageLoadOptions{
		Children:                true,
		HasGrandChildren:        true,
		RedLinkCountForChildren: true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(data.ParentID, returnData.PageMap, loadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}
	// Remove parent, since we only want to return children.
	delete(returnData.PageMap, data.ParentID)

	return pages.Success(returnData)
}
