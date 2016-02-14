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

// newLikeHandlerFunc handles requests to create/update a prior like.
func newLikeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	var task newLikeData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&task)
	if err != nil || !core.IsIdValid(task.PageId) {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if task.Value < -1 || task.Value > 1 {
		return pages.HandlerBadRequestFail("Value has to be -1, 0, or 1", nil)
	}

	// Check to see if we have a recent like by this user for this page.
	var id string
	var found bool
	row := db.NewStatement(`
		SELECT id
		FROM likes
		WHERE userId=? AND pageId=? AND TIME_TO_SEC(TIMEDIFF(?, createdAt)) < ?
		`).QueryRow(u.Id, task.PageId, database.Now(), redoWindow)
	found, err = row.Scan(&id)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a recent like", err)
	}
	if found {
		hashmap := make(map[string]interface{})
		hashmap["id"] = id
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("likes", hashmap, "value", "createdAt")
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't update a like", err)
		}
	} else {
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["pageId"] = task.PageId
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("likes", hashmap)
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't add a like", err)
		}
	}
	return pages.StatusOK(nil)
}
