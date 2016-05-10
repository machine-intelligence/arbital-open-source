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
	returnData := core.NewHandlerData(u).SetResetEverything()

	return pages.StatusOK(returnData)
}
