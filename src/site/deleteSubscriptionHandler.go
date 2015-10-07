// deleteSubscriptionHandler.go handles requests for deleting a subscription.
package site

import (
	"encoding/json"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// deleteSubscriptionData contains the data we receive in the request.
type deleteSubscriptionData struct {
	PageId int64 `json:",string"`
	UserId int64 `json:",string"`
}

// deleteSubscriptionHandler handles requests for deleting a subscription.
func deleteSubscriptionHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	decoder := json.NewDecoder(params.R.Body)
	var data deleteSubscriptionData
	err := decoder.Decode(&data)
	if err != nil || (data.PageId == 0 && data.UserId == 0) {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	query := database.NewQuery(`
		DELETE FROM subscriptions
		WHERE userId=? AND `, u.Id)
	if data.PageId > 0 {
		query.Add("toPageId=?", data.PageId)
	} else if data.UserId > 0 {
		query.Add("toUserId=?", data.UserId)
	}
	if _, err := query.ToStatement(db).Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't delete a subscription", err)
	}
	return pages.StatusOK(nil)
}
