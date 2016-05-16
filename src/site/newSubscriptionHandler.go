// newSubscription.go handles repages for adding a new subscription.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// newSubscriptionData contains the data we get in the request.
type newSubscriptionData struct {
	PageId string
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("Page id has to be set", err).Status(http.StatusBadRequest)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return addSubscription(tx, u.Id, data.PageId)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}
	return pages.Success(nil)
}

func addSubscription(tx *database.Tx, userId string, toPageId string) sessions.Error {
	hashmap := make(map[string]interface{})
	hashmap["userId"] = userId
	hashmap["toId"] = toPageId
	hashmap["createdAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("subscriptions", hashmap, "userId").WithTx(tx)
	_, err := statement.Exec()
	if err != nil {
		return sessions.NewError("Couldn't subscribe", err)
	}
	return nil
}
