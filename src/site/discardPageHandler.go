// discardPageHandler.go handles requests for discarding a page. This means
// deleting all autosaves which were created by the user.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// discardPageData is the data received from the request.
type discardPageData struct {
	PageID string
}

var discardPageHandler = siteHandler{
	URI:         "/discardPage/",
	HandlerFunc: discardPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// discardPageHandlerFunc handles requests for deleting a page.
func discardPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data discardPageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Delete the edit
	statement := db.NewStatement(`
		DELETE FROM pages
		WHERE pageId=? AND creatorId=? AND isAutosave`)
	if _, err = statement.Exec(data.PageID, u.ID); err != nil {
		return pages.Fail("Couldn't discard a page", err)
	}

	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageID
	hashmap["lockedUntil"] = database.Now()
	statement = db.NewInsertStatement("pageInfos", hashmap, "lockedUntil")
	if _, err = statement.Exec(); err != nil {
		return pages.Fail("Couldn't change lock", err)
	}
	return pages.Success(nil)
}
