// readModePageJsonHandler.go serves the /read page.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var readModePageJsonHandler = siteHandler{
	URI:         "/json/readMode/",
	HandlerFunc: readModePageHandlerFunc,
	Options:     pages.PageOptions{},
}

type HotPageData struct {
	PageId    string `json:"pageId"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

func readModePageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// figure out which pages are to show as exciting and hot!
	hotPages, err := getHotPages(db, u)
	if err != nil {
		return pages.HandlerErrorFail("failed to load hot page ids", err)
	}

	// add the hot pages to the map
	hotPageLoadOptions := (&core.PageLoadOptions{Creators: true}).Add(core.TitlePlusLoadOptions)
	for _, data := range hotPages {
		core.AddPageToMap(data.PageId, returnData.PageMap, hotPageLoadOptions)
	}

	// load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["hotPages"] = hotPages
	return pages.StatusOK(returnData)
}

func getHotPages(db *database.DB, u *core.CurrentUser) ([]HotPageData, error) {
	var hotPages []HotPageData

	rows := database.NewQuery(`
		SELECT pageId, createdBy, createdAt
		FROM`).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.type IN ('wiki', 'lens', 'domain', 'question')
		ORDER BY createdAt DESC
		LIMIT 100`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var data HotPageData
		err := rows.Scan(&data.PageId, &data.CreatedBy, &data.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan for hot pages: %v", err)
		}
		hotPages = append(hotPages, data)
		return nil
	})

	return hotPages, err
}
