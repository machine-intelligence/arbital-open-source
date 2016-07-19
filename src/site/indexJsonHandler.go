// indexJsonHandler.go serves the index page data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type featuredDomain struct {
	DomainID string   `json:"domainId"`
	ChildIds []string `json:"childIds"`
}

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
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
