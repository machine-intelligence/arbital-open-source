// newSubscription.go handles repages for adding a new subscription.
package site

import (
	"encoding/json"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newSubscriptionData contains the data we get in the request.
type newSubscriptionData struct {
	PageId int64 `json:",string"`
	UserId int64 `json:",string"`
}

// newSubscriptionHandler handles requests for adding a new subscription.
func newSubscriptionHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data newSubscriptionData
	err := decoder.Decode(&data)
	if err != nil || (data.PageId == 0 && data.UserId == 0) {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	if data.PageId > 0 {
		err = addSubscriptionToPage(db, u.Id, data.PageId)
	} else if data.UserId > 0 {
		err = addSubscriptionToUser(db, u.Id, data.UserId)
	}
	if err != nil {
		return pages.HandlerErrorFail("Couldn't create new subscription", err)
	}
	return pages.StatusOK(nil)
}

func addSubscriptionToPage(db *database.DB, userId int64, pageId int64) error {
	hashmap := map[string]interface{}{"toPageId": pageId}
	return addSubscription(db, hashmap, userId)
}

func addSubscriptionToUser(db *database.DB, userId int64, toUserId int64) error {
	hashmap := map[string]interface{}{"toUserId": toUserId}
	return addSubscription(db, hashmap, userId)
}

func addSubscription(db *database.DB, hashmap map[string]interface{}, userId int64) error {
	hashmap["userId"] = userId
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("subscriptions", hashmap, "userId")
	_, err := statement.Exec()
	return err
}
