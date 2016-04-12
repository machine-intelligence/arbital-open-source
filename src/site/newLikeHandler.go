// newLikeHandler.go adds a new like for for a page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	redoWindow = 60 // number of seconds during which a user can redo their like
)

// newLikeData contains data given to us in the request.
type newLikeData struct {
	PageId string
	Value  int
}

var newLikeHandler = siteHandler{
	URI:         "/newLike/",
	HandlerFunc: newLikeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newLikeHandlerFunc handles requests to create/update a like.
func newLikeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var data newLikeData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if data.Value < -1 || data.Value > 1 {
		return pages.HandlerBadRequestFail("Value has to be -1, 0, or 1", nil)
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["pageId"] = data.PageId
	hashmap["value"] = data.Value
	hashmap["updatedAt"] = database.Now()
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("likes", hashmap, "value", "updatedAt")
	if _, err = statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't update a like", err)
	}

	return pages.StatusOK(nil)
}
