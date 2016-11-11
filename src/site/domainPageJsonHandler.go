// domainPageJsonHandler.go serves data to display domain index page.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type domainPageData struct {
	DomainAlias string
}

var domainPageHandler = siteHandler{
	URI:         "/json/domainPage/",
	HandlerFunc: domainPageHandlerFunc,
	Options:     pages.PageOptions{},
}

// domainPageJsonHandler handles the request.
func domainPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data domainPageData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	returnData.ResultMap["domain"], err = core.LoadDomain(db, data.DomainAlias)
	if err != nil {
		return pages.Fail("Couldn't load domain", err)
	}

	return pages.Success(returnData)
}
