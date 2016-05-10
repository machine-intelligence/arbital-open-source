// readModePageJsonHandler.go serves the /read page.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var readModePageJsonHandler = siteHandler{
	URI:         "/json/readMode/",
	HandlerFunc: readModePageHandlerFunc,
	Options:     pages.PageOptions{},
}

func readModePageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// figure out which pages are to show as exciting and hot!
	hotPageIds, err := loadHotPageIds(db, u, returnData.PageMap)
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

func loadHotPageIds(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page) ([]string, error) {
	rows := database.NewQuery(`
		SELECT pageId
		FROM`).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.type IN ('wiki', 'lens', 'domain', 'question')
		ORDER BY createdAt DESC
		LIMIT 100`).ToStatement(db).Query()

	pageOptions := (&core.PageLoadOptions{SubpageCounts: true}).Add(core.TitlePlusLoadOptions)
	return core.LoadPageIds(rows, pageMap, pageOptions)
}
