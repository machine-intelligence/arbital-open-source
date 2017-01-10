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
	Table        string `json:"table"`
	ToID         string `json:"toId"`
	IsSubscribed bool   `json:"isSubscribed"`
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
	if !core.IsIDValid(data.ToID) {
		return pages.Fail("ToId has to be set", err).Status(http.StatusBadRequest)
	}
	if data.Table != core.DiscussionSubscriptionTable &&
		data.Table != core.UserSubscriptionTable &&
		data.Table != core.MaintainerSubscriptionTable {
		return pages.Fail("Invalid subscription table", nil).Status(http.StatusBadRequest)
	}

	if !data.IsSubscribed {
		// Delete the subscription
		toId := `toPageId`
		if data.Table == core.UserSubscriptionTable {
			toId = `toUserId`
		}
		query := database.NewQuery(`
			DELETE FROM `+data.Table+`
			WHERE userId=?`, u.ID).Add(`
				AND `+toId+`=?`, data.ToID)
		if _, err := query.ToStatement(db).Exec(); err != nil {
			return pages.Fail("Couldn't delete a subscription", err)
		}
		return pages.Success(nil)
	}

	// Otherwise, create/update it
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return core.AddSubscription(tx, u.ID, data.Table, data.ToID)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}
	return pages.Success(nil)
}
