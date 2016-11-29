// settingsPageJsonHandler.go contains the handler for returning JSON with data
// to display the settings/invite page.

package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type Domain struct {
	DomainID string `json:"domainId"`
	LongName string `json:"longName"`
}

var settingsPageHandler = siteHandler{
	URI:         "/json/settingsPage/",
	HandlerFunc: settingsPageJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// settingsPageJsonHandler renders the settings page.
func settingsPageJSONHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
