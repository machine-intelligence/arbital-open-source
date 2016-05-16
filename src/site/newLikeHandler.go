// newLikeHandler.go adds a new like for for a page.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

const (
	redoWindow = 60 // number of seconds during which a user can redo their like
)

// newLikeData contains data given to us in the request.
type newLikeData struct {
	LikeableType string // Eg 'changelog', 'page'
	Id           string
	Value        int
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if data.LikeableType == PageLikeableType && !core.IsIdValid(data.Id) {
		return pages.Fail("Invalid page id", nil).Status(http.StatusBadRequest)
	}
	if data.Value < -1 || data.Value > 1 {
		return pages.Fail("Value has to be -1, 0, or 1", nil).Status(http.StatusBadRequest)
	}
	if !(data.LikeableType == ChangelogLikeableType || data.LikeableType == PageLikeableType) {
		return pages.Fail("LikeableType has to be 'changelog' or 'page'", nil).Status(http.StatusBadRequest)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Snapshot the state of the user's trust.
		snapshotId, err := InsertUserTrustSnapshots(tx, u)
		if err != nil {
			return sessions.NewError("Couldn't insert userTrustSnapshot", err)
		}

		// Get the likeableId of this likeable.
		likeableId, err := GetOrCreateLikeableId(tx, data.LikeableType, data.Id)
		if err != nil {
			return sessions.NewError("Couldn't get the likeableId", err)
		}

		// Create/update the like.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["likeableId"] = likeableId
		hashmap["value"] = data.Value
		hashmap["createdAt"] = database.Now()
		hashmap["updatedAt"] = database.Now()
		hashmap["userTrustSnapshotId"] = snapshotId
		statement := db.NewInsertStatement("likes", hashmap, "value", "updatedAt", "userTrustSnapshotId")
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return sessions.NewError("Couldn't update/create a like", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
