// discardPageHandler.go handles requests for discarding a page. This means
// deleting all autosaves which were created by the user.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// discardPageData is the data received from the request.
type discardPageData struct {
	PageId string `json:""`
}

var discardPageHandler = siteHandler{
	URI:         "/discardPage/",
	HandlerFunc: discardPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// discardPageHandlerFunc handles requests for deleting a page.
func discardPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data discardPageData
	err := decoder.Decode(&data)
	if err != nil || !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Delete the edit
	statement := db.NewStatement(`
		DELETE FROM pages
		WHERE pageId=? AND creatorId=? AND isAutosave`)
	if _, err = statement.Exec(data.PageId, u.Id); err != nil {
		return pages.HandlerErrorFail("Couldn't discard a page", err)
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["lockedUntil"] = database.Now()
	statement = db.NewInsertStatement("pageInfos", hashmap, "lockedUntil")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't change lock", err)
	}
	return pages.StatusOK(nil)
}
