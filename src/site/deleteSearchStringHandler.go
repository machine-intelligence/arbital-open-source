// deleteSearchStringHandler.go adds a search string to a page
package site

import (
	"encoding/json"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// deleteSearchStringData contains data given to us in the request.
type deleteSearchStringData struct {
	Id string
}

var deleteSearchStringHandler = siteHandler{
	URI:         "/deleteSearchString/",
	HandlerFunc: deleteSearchStringHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// deleteSearchStringHandlerFunc handles requests to create/update a like.
func deleteSearchStringHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	var data deleteSearchStringData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	id, err := strconv.ParseInt(data.Id, 10, 64)
	if err != nil {
		return pages.HandlerBadRequestFail("Invalid id", err)
	}

	statement := database.NewQuery(`
		DELETE FROM searchStrings WHERE id=?`, id).ToStatement(db)
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't delete from DB", err)
	}

	return pages.StatusOK(nil)
}
