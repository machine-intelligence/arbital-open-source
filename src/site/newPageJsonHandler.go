// newPageJsonHandler.go creates and returns a new page
package site

import (
	"math/rand"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// newPageJsonHandler handles the request.
func newPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	if !params.U.IsLoggedIn {
		return pages.HandlerBadRequestFail("Have to be logged in", nil)
	}

	pageMap := make(map[int64]*core.Page)
	core.AddPageIdToMap(rand.Int63(), pageMap)

	returnData := createReturnData(pageMap)
	return pages.StatusOK(returnData)
}
