// newSearchStringHandler.go adds a search string to a page
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newSearchStringData contains data given to us in the request.
type newSearchStringData struct {
	PageId string
	Text   string
}

var newSearchStringHandler = siteHandler{
	URI:         "/newSearchString/",
	HandlerFunc: newSearchStringHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newSearchStringHandlerFunc handles requests to create/update a like.
func newSearchStringHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, false)

	var data newSearchStringData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if len(data.Text) <= 0 {
		return pages.HandlerBadRequestFail("Invalid text", nil)
	}

	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["text"] = data.Text
	hashmap["userId"] = u.Id
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("searchStrings", hashmap)
	resp, err := statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't insert into DB", err)
	}

	newId, err := resp.LastInsertId()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't get inserted id", err)
	}

	returnData.ResultMap["searchStringId"] = fmt.Sprintf("%d", newId)
	return pages.StatusOK(returnData.ToJson())
}
