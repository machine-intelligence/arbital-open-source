// redirectToPrimaryPageJsonHandler.go contains the handler for returning JSON with data
// to redirect to a primary page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// redirectToPrimaryPageJsonData contains parameters passed in via the request.
type redirectToPrimaryPageJsonData struct {
	PageAlias string
}

var redirectToPrimaryPageHandler = siteHandler{
	URI:         "/json/redirectToPrimaryPage/",
	HandlerFunc: redirectToPrimaryPageJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

// redirectToPrimaryPageJsonHandler handles the request.
func redirectToPrimaryPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Decode data
	var data redirectToPrimaryPageJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	pageId, ok, err := core.LoadOldAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	return pages.StatusOK(pageId)
}
