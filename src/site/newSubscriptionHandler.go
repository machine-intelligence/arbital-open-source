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
}

var newSubscriptionHandler = siteHandler{
	URI:         "/newSubscription/",
	HandlerFunc: newSubscriptionHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newSubscriptionHandlerFunc handles requests for adding a new subscription.
func newSubscriptionHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data newSubscriptionData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.PageId <= 0 {
		return pages.HandlerBadRequestFail("Page id has to be set", err)
	}

	errorMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		return addSubscription(tx, u.Id, data.PageId)
	})
	if errorMessage != "" {
		return pages.HandlerErrorFail(errorMessage, err)
	}
	return pages.StatusOK(nil)
}

func addSubscription(tx *database.Tx, userId int64, toPageId int64) (string, error) {
	hashmap := make(map[string]interface{})
	hashmap["userId"] = userId
	hashmap["toId"] = toPageId
	hashmap["createdAt"] = database.Now()
	statement := tx.NewInsertTxStatement("subscriptions", hashmap, "userId")
	_, err := statement.Exec()
	if err != nil {
		return "Couldn't subscribe", err
	}
	return "", nil
}
