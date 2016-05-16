// deleteSubscriptionHandler.go handles requests for deleting a subscription.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// deleteSubscriptionData contains the data we receive in the request.
type deleteSubscriptionData struct {
	PageId string
}

var deleteSubscriptionHandler = siteHandler{
	URI:         "/deleteSubscription/",
	HandlerFunc: deleteSubscriptionHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// deleteSubscriptionHandlerFunc handles requests for deleting a subscription.
func deleteSubscriptionHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Get and check data
	var data deleteSubscriptionData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Page id has to be set", err)
	}

	query := database.NewQuery(`
		DELETE FROM subscriptions
		WHERE userId=? AND toId=?`, u.Id, data.PageId)
	if _, err := query.ToStatement(db).Exec(); err != nil {
		return pages.Fail("Couldn't delete a subscription", err)
	}
	return pages.Success(nil)
}
