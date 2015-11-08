// defaultJsonHandler.go returns basic data every page needs. Used for pages
// that don't need any custom data, and therefore don't have custom handlers.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var defaultHandler = siteHandler{
	URI:         "/json/default/",
	HandlerFunc: defaultJsonHandlerFunc,
}

func defaultJsonHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	returnData := newHandlerData(true)
	returnData.User = u
	err := core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
