// deleteSubscriptionHandler.go handles requests for deleting a subscription.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// deleteSubscriptionData contains the data we receive in the request.
type deleteSubscriptionData struct {
	PageId int64 `json:",string"`
	UserId int64 `json:",string"`
}

// deleteSubscriptionHandler handles requests for deleting a subscription.
func deleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var emptyData, data deleteSubscriptionData
	err := decoder.Decode(&data)
	if err != nil || data == emptyData {
		c.Inc("delete_subscription_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("delete_subscription_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	query := fmt.Sprintf(`
		DELETE FROM subscriptions
		WHERE userId=%d AND `, u.Id)
	if data.PageId > 0 {
		query += fmt.Sprintf("toPageId=%d", data.PageId)
	} else if data.UserId > 0 {
		query += fmt.Sprintf("toUserId=%d", data.UserId)
	}
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("delete_subscription_fail")
		c.Errorf("Couldn't delete a subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
