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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.LikeableType == PageLikeableType && !core.IsIdValid(data.Id) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if data.Value < -1 || data.Value > 1 {
		return pages.HandlerBadRequestFail("Value has to be -1, 0, or 1", nil)
	}
	if !(data.LikeableType == ChangelogLikeableType || data.LikeableType == PageLikeableType) {
		return pages.HandlerBadRequestFail("LikeableType has to be 'changelog' or 'page'", nil)
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Snapshot the state of the user's trust.
		snapshotId, err := InsertUserTrustSnapshots(tx, u)
		if err != nil {
			return "Couldn't insert userTrustSnapshot", err
		}

		// Get the likeableId of this likeable.
		likeableId, err := GetOrCreateLikeableId(tx, data.LikeableType, data.Id)
		if err != nil {
			return "Couldn't get the likeableId", err
		}

		// Create/update the like.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["likeableId"] = likeableId
		hashmap["value"] = data.Value
		hashmap["createdAt"] = database.Now()
		hashmap["updatedAt"] = database.Now()
		hashmap["userTrustSnapshotId"] = snapshotId
		params.C.Debugf("================ %+v", hashmap)
		statement := db.NewInsertStatement("likes", hashmap, "value", "updatedAt", "userTrustSnapshotId")
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return "Couldn't update/create a like", err
		}

		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail("Couldn't insert a like: "+errMessage, err)
	}

	return pages.StatusOK(nil)
}
