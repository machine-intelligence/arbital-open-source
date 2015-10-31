// newPageJsonHandler.go creates and returns a new page
package site

import (
	"math/rand"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var newPageHandler = siteHandler{
	URI:         "/json/newPage/",
	HandlerFunc: newPageJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newPageJsonHandler handles the request.
func newPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	returnData := newHandlerData()
	core.AddPageIdToMap(rand.Int63(), returnData.PageMap)
	return pages.StatusOK(returnData.toJson())
}
