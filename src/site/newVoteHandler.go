// newVoteHandler.go adds a new vote for for a page.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	redoWindow = 60 // number of seconds during which a user can redo their like
)

// newVoteData contains data given to us in the request.
type newVoteData struct {
	PageID string
	Value  float32
}

var newVoteHandler = siteHandler{
	URI:         "/newVote/",
	HandlerFunc: newVoteHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newVoteHandlerFunc handles requests to create/update a prior vote.
func newVoteHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var task newVoteData
	err := decoder.Decode(&task)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(task.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}
	if task.Value < -1 || task.Value > 100 {
		return pages.Fail("Value has to be mu or [0, 100]", nil).Status(http.StatusBadRequest)
	}

	// Get the last vote.
	var oldVoteID int64
	var oldVoteValue float32
	var oldVoteExists bool
	var oldVoteAge int64
	row := db.NewStatement(`
		SELECT id,value,TIME_TO_SEC(TIMEDIFF(?,createdAt)) AS age
		FROM votes
		WHERE userId=? AND pageId=?
		ORDER BY id DESC
		LIMIT 1`).QueryRow(database.Now(), u.ID, task.PageID)
	oldVoteExists, err = row.Scan(&oldVoteID, &oldVoteValue, &oldVoteAge)
	if err != nil {
		return pages.Fail("Couldn't check for a recent vote", err)
	}

	// If previous vote is exactly the same, don't do anything.
	if oldVoteExists && oldVoteValue == task.Value {
		return pages.Success(nil)
	}

	// Check to see if we have a recent vote by this user for this page.
	if oldVoteExists && oldVoteAge <= redoWindow {
		hashmap := make(map[string]interface{})
		hashmap["id"] = oldVoteID
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("votes", hashmap, "value", "createdAt")
		if _, err = statement.Exec(); err != nil {
			return pages.Fail("Couldn't update a vote", err)
		}
	} else {
		// Insert new vote.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.ID
		hashmap["pageId"] = task.PageID
		hashmap["value"] = task.Value
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("votes", hashmap)
		if _, err = statement.Exec(); err != nil {
			return pages.Fail("Couldn't add a vote", err)
		}
	}
	return pages.Success(nil)
}
