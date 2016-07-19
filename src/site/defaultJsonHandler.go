// defaultJsonHandler.go returns basic data every page needs. Used for pages
// that don't need any custom data, and therefore don't have custom handlers.

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var defaultHandler = siteHandler{
	URI:         "/json/default/",
	HandlerFunc: defaultJSONHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}

func defaultJSONHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	returnData := core.NewHandlerData(params.U).SetResetEverything()
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
