// externalUrlHandler.go gets info about an external url

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var externalUrlHandler = siteHandler{
	URI:         "/isDuplicateExternalUrl/",
	HandlerFunc: externalUrlHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// externalUrlData contains parameters passed in via the request.
type externalUrlData struct {
	ExternalUrl string
}

// externalUrlHandlerFunc handles the request.
func externalUrlHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	// Decode data
	var data externalUrlData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	isDupe, originalPageID, err := core.IsDuplicateExternalUrl(db, u, data.ExternalUrl)
	if err != nil {
		return pages.Fail("Couldn't check if external url is already in use.", err)
	}
	returnData.ResultMap["isDupe"] = isDupe
	returnData.ResultMap["originalPageID"] = originalPageID

	// Load data
	core.AddPageToMap(originalPageID, returnData.PageMap, core.TitlePlusLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
