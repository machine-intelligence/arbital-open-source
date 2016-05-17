// readModeHandler.go serves the /read panel.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type readModeData struct {
	NumPagesToLoad int
}

var readModeHandler = siteHandler{
	URI:         "/json/readMode/",
	HandlerFunc: readModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func readModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data readModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// figure out which pages to show as exciting and hot!
	numPagesToLoad := data.NumPagesToLoad
	if numPagesToLoad == 0 {
		numPagesToLoad = 25
	}
	hotPageIds, err := loadHotPageIds(db, u, returnData.PageMap, numPagesToLoad)
	if err != nil {
		return pages.Fail("failed to load hot page ids", err)
	}

	// Load and update LastReadModeView for this user
	returnData.ResultMap[LastReadModeView], err = LoadAndUpdateLastView(db, u, LastReadModeView)
	if err != nil {
		return pages.Fail("Error updating last read mode view", err)
	}

	// Uncomment this to test the feature.
	// returnData.ResultMap[LastReadModeView] = "2016-05-14 20:11:42"

	// load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["hotPageIds"] = hotPageIds
	return pages.Success(returnData)
}

func loadHotPageIds(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, numPagesToLoad int) ([]string, error) {
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.type IN (?,?,?,?)`, core.WikiPageType, core.LensPageType, core.DomainPageType, core.QuestionPageType).Add(`
		ORDER BY createdAt DESC
		LIMIT ?`, numPagesToLoad).ToStatement(db).Query()

	pageOptions := (&core.PageLoadOptions{SubpageCounts: true}).Add(core.TitlePlusLoadOptions)
	return core.LoadPageIds(rows, pageMap, pageOptions)
}
