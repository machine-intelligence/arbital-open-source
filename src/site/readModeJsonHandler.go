// readModeJsonHandler.go serves the /read panel.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type readModeJsonData struct {
	NumPagesToLoad int
}

var readModeJsonHandler = siteHandler{
	URI:         "/json/readMode/",
	HandlerFunc: readModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func readModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data readModeJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// figure out which pages to show as exciting and hot!
	numPagesToLoad := data.NumPagesToLoad
	if numPagesToLoad == 0 {
		numPagesToLoad = 25
	}
	hotPageIds, err := loadHotPageIds(db, u, returnData.PageMap, numPagesToLoad)
	if err != nil {
		return pages.HandlerErrorFail("failed to load hot page ids", err)
	}

	// load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["hotPageIds"] = hotPageIds
	return pages.StatusOK(returnData)
}

func loadHotPageIds(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, numPagesToLoad int) ([]string, error) {
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.type IN ('wiki', 'lens', 'domain', 'question')
		ORDER BY createdAt DESC
		LIMIT ?`, numPagesToLoad).ToStatement(db).Query()

	pageOptions := (&core.PageLoadOptions{SubpageCounts: true}).Add(core.TitlePlusLoadOptions)
	return core.LoadPageIds(rows, pageMap, pageOptions)
}
