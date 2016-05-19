// Handles requests to update subscriptions
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

type updateSubscriptionData struct {
	ToId         string `json:"toId"`
	IsSubscribed bool   `json:"isSubscribed"`
	AsMaintainer bool   `json:asMaintainer"`
}

var updateSubscriptionHandler = siteHandler{
	URI:         "/updateSubscription/",
	HandlerFunc: updateSubscriptionHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateSubscriptionHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data updateSubscriptionData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.ToId) {
		return pages.Fail("ToId has to be set", err).Status(http.StatusBadRequest)
	}

	if !data.IsSubscribed {
		// Delete the subscription
		query := database.NewQuery(`
			DELETE FROM subscriptions
			WHERE userId=? AND toId=?`, u.Id, data.ToId)
		if _, err := query.ToStatement(db).Exec(); err != nil {
			return pages.Fail("Couldn't delete a subscription", err)
		}
		return pages.Success(nil)
	}

	// Otherwise, create/update it
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return addSubscription(tx, u.Id, data.ToId, data.AsMaintainer)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}
	return pages.Success(nil)
}

func addSubscription(tx *database.Tx, userId string, toPageId string, asMaintainer bool) sessions.Error {
	hashmap := make(map[string]interface{})
	hashmap["userId"] = userId
	hashmap["toId"] = toPageId
	hashmap["createdAt"] = database.Now()
	hashmap["asMaintainer"] = asMaintainer
	statement := tx.DB.NewInsertStatement("subscriptions", hashmap, "asMaintainer").WithTx(tx)
	_, err := statement.Exec()
	if err != nil {
		return sessions.NewError("Couldn't subscribe", err)
	}
	return nil
}
