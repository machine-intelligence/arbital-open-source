// domainsPageJsonHandler.go serves the domains data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var domainsPageHandler = siteHandler{
	URI:         "/json/domainsPage/",
	HandlerFunc: domainsPageHandlerFunc,
	Options: pages.PageOptions{
		AdminOnly: true,
	},
}

func domainsPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var err error
	returnData := core.NewHandlerData(params.U, true)

	// Load the domains
	err = core.LoadDomainIds(db, u, nil, returnData.PageMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading group members", err)
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.ToJson())
}
