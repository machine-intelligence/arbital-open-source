// newVoteHandler.go adds a new vote for for a page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newVoteData contains data given to us in the request.
type newVoteData struct {
	PageId string `json:""`
	Value  float32
}

var newVoteHandler = siteHandler{
	URI:         "/newVote/",
	HandlerFunc: newVoteHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// newVoteHandlerFunc handles requests to create/update a prior vote.
func newVoteHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var task newVoteData
	err := decoder.Decode(&task)
	if err != nil || !core.IsIdValid(task.PageId) {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if task.Value < 0 || task.Value > 100 {
		return pages.HandlerBadRequestFail("Value has to be [0, 100]", nil)
	}

	// Get the last vote.
	var oldVoteId int64
	var oldVoteValue float32
	var oldVoteExists bool
	var oldVoteAge int64
	row := db.NewStatement(`
		SELECT id,value,TIME_TO_SEC(TIMEDIFF(?,createdAt)) AS age
		FROM votes
		WHERE userId=? AND pageId=?
		ORDER BY id DESC
		LIMIT 1`).QueryRow(database.Now(), u.Id, task.PageId)
	oldVoteExists, err = row.Scan(&oldVoteId, &oldVoteValue, &oldVoteAge)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't check for a recent vote", err)
	}

	// If previous vote is exactly the same, don't do anything.
	if oldVoteExists && oldVoteValue == task.Value {
		return pages.StatusOK(nil)
	}

	// Check to see if we have a recent vote by this user for this page.
	if oldVoteExists && oldVoteAge <= redoWindow {
		hashmap := make(map[string]interface{})
		hashmap["id"] = oldVoteId
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("votes", hashmap, "value", "createdAt")
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't update a vote", err)
		}
	} else {
		// Insert new vote.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["pageId"] = task.PageId
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("votes", hashmap)
		if _, err = statement.Exec(); err != nil {
			return pages.HandlerErrorFail("Couldn't add a vote", err)
		}
	}
	return pages.StatusOK(nil)
}
