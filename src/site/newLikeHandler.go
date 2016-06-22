// newLikeHandler.go adds a new (dis)like for a likeable object
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// newLikeData contains data given to us in the request.
type newLikeData struct {
	// If likeableId is given, we'll modify the corresponding likeable directly, ignoring objectId
	LikeableId int64 `json:"likeableId,string"`

	// If likeableId is missing, we'll use objectId to look it up in the appropriate
	// table (based on likeableType)
	// For example, if likeableType=='page', the objectId is the corresponding pageId
	ObjectId     string
	LikeableType string // Eg 'changelog', 'page'

	Value int
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
	if data.Value < -1 || data.Value > 1 {
		return pages.Fail("Value has to be -1, 0, or 1", nil).Status(http.StatusBadRequest)
	}
	if !core.IsValidLikeableType(data.LikeableType) {
		return pages.Fail("LikeableType is not valid", nil).Status(http.StatusBadRequest)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		if data.LikeableId == 0 {
			// Get the likeableId of this likeable.
			data.LikeableId, err = core.GetOrCreateLikeableId(tx, data.LikeableType, data.ObjectId)
			if err != nil {
				return sessions.NewError("Couldn't get the likeableId", err)
			}
		}

		// Snapshot the state of the user's trust.
		var snapshotId int64
		if data.LikeableType == core.ChangeLogLikeableType || data.LikeableType == core.PageLikeableType {
			snapshotId, err = InsertUserTrustSnapshots(tx, u)
			if err != nil {
				return sessions.NewError("Couldn't insert userTrustSnapshot", err)
			}
		}

		// Create/update the like.
		hashmap := make(map[string]interface{})
		hashmap["userId"] = u.Id
		hashmap["likeableId"] = data.LikeableId
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
