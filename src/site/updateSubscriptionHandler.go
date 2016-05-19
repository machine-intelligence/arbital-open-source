// Handles requests to update subscriptions
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

type updateSubscriptionData struct {
	ToId         string `json:"toId"`
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

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return addSubscription(tx, u.Id, data.ToId, data.AsMaintainer)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
