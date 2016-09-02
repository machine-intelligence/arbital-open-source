// indexJsonHandler.go serves the index page data.

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var indexHandler = siteHandler{
	URI:         "/json/index/",
	HandlerFunc: indexJSONHandler,
	Options:     pages.PageOptions{},
}

func indexJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Load pages.
	core.AddPageIDToMap("3hs", returnData.PageMap)
	core.AddPageIDToMap("4ym", returnData.PageMap)
	core.AddPageToMap("4c7", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("600", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("5xx", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("1ln", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("5zs", returnData.PageMap, core.TitlePlusLoadOptions)
	core.AddPageToMap("3rb", returnData.PageMap, core.TitlePlusLoadOptions)
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
