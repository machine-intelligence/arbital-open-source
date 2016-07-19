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
	LikeableID int64 `json:"likeableId,string"`

	// If likeableId is missing, we'll use objectId to look it up in the appropriate
	// table (based on likeableType)
	// For example, if likeableType=='page', the objectId is the corresponding pageId
	ObjectID     string
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
		return addNewLike(tx, u, data.LikeableID, data.ObjectID, data.LikeableType, data.Value)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}

func addNewLike(tx *database.Tx, u *core.CurrentUser, likeableID int64, objectID string, likeableType string, value int) sessions.Error {
	var err error

	if likeableID == 0 {
		// Get the likeableId of this likeable.
		likeableID, err = core.GetOrCreateLikeableID(tx, likeableType, objectID)
		if err != nil {
			return sessions.NewError("Couldn't get the likeableId", err)
		}
	}

	// Snapshot the state of the user's trust.
	var snapshotID int64
	if likeableType == core.ChangeLogLikeableType || likeableType == core.PageLikeableType {
		snapshotID, err = InsertUserTrustSnapshots(tx, u)
		if err != nil {
			return sessions.NewError("Couldn't insert userTrustSnapshot", err)
		}
	}

	// Create/update the like.
	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.ID
	hashmap["likeableId"] = likeableID
	hashmap["value"] = value
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	hashmap["userTrustSnapshotId"] = snapshotID
	statement := tx.DB.NewInsertStatement("likes", hashmap, "value", "updatedAt", "userTrustSnapshotId")
	if _, err = statement.WithTx(tx).Exec(); err != nil {
		return sessions.NewError("Couldn't update/create a like", err)
	}

	return nil
}
