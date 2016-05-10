// hotStuffHandler.go serves the hot stuff page.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var hotStuffHandler = siteHandler{
	URI:         "/json/hotstuff/",
	HandlerFunc: hotStuffHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}

func hotStuffHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// figure out which pages are to show as exciting and hot!
	hotPageIds, err := getHotPageIds(db, u)
	if err != nil {
		return pages.HandlerErrorFail("failed to load hot page ids", err)
	}

	// add the hot pages to the map
	for _, pageId := range hotPageIds {
		core.AddPageToMap(pageId, returnData.PageMap, core.TitlePlusLoadOptions)
	}

	// load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["hotPageIds"] = hotPageIds
	return pages.StatusOK(returnData)
}

func getHotPageIds(db *database.DB, u *core.CurrentUser) ([]string, error) {
	var hotPageIds []string

	rows := database.NewQuery(`
		SELECT
			pageId
		FROM`).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.type IN ('wiki', 'lens', 'domain', 'question')
		ORDER BY createdAt DESC
		LIMIT 100`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("failed to scan for hot pages: %v", err)
		}
		hotPageIds = append(hotPageIds, pageId)
		return nil
	})

	return hotPageIds, err
}
