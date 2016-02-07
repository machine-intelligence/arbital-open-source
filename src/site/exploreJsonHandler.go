// explorePage.go serves the explore template.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type exploreJsonData struct {
	GroupAlias string
}

var exploreHandler = siteHandler{
	URI:         "/json/explore/",
	HandlerFunc: exploreJsonHandler,
}

func exploreJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	// Decode data
	var data exploreJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual domain id
	var domainId string
	if data.GroupAlias != "" {
		var ok bool
		var err error
		domainId, ok, err = core.LoadAliasToPageId(db, data.GroupAlias)
		if err != nil {
			return pages.Fail("Couldn't convert alias", err)
		}
		if !ok {
			return pages.Fail("Couldn't find alias", nil)
		}
	} else if core.IsIdValid(params.PrivateGroupId) {
		domainId = params.PrivateGroupId
	} else {
		return pages.HandlerBadRequestFail("No domain specified", nil)
	}

	returnData := newHandlerData(true)
	returnData.User = u
	returnData.ResultMap["rootPageId"] = domainId

	// Load the root page
	loadOptions := (&core.PageLoadOptions{
		Children:                true,
		HasGrandChildren:        true,
		RedLinkCountForChildren: true,
		RedLinkCount:            true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(domainId, returnData.PageMap, loadOptions)

	// Load pages.
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
