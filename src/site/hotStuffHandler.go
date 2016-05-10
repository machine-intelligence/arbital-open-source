// hotStuffHandler.go serves the hot stuff page.
package site

import (
	"zanaduu3/src/core"
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

	hotPageIds := []string{"1y", "11x"}
	returnData.ResultMap["hotPageIds"] = hotPageIds

	for _, pageId := range hotPageIds {
		core.AddPageToMap(pageId, returnData.PageMap, core.TitlePlusLoadOptions)
	}

	// Load pages.
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}
