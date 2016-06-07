// indexJsonHandler.go serves the index page data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type featuredDomain struct {
	DomainId string   `json:"domainId"`
	ChildIds []string `json:"childIds"`
}

var indexHandler = siteHandler{
	URI:         "/json/index/",
	HandlerFunc: indexJsonHandler,
	Options:     pages.PageOptions{},
}

func indexJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Load pages.
	core.AddPageIdToMap("3d", returnData.PageMap)
	core.AddPageIdToMap("1sl", returnData.PageMap)
	core.AddPageIdToMap("1sm", returnData.PageMap)
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
