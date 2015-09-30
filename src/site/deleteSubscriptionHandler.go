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

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("delete_subscription_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
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

	err = deleteSubscriptionInternalHandler(u, db, &data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.Inc("delete_subscription_fail")
	}
}

func deleteSubscriptionInternalHandler(u *user.User, db *database.DB, data *deleteSubscriptionData) error {
	query := database.NewQuery(`
		DELETE FROM subscriptions
		WHERE userId=? AND `, u.Id)
	if data.PageId > 0 {
		query.Add("toPageId=?", data.PageId)
	} else if data.UserId > 0 {
		query.Add("toUserId=?", data.UserId)
	}
	if _, err := query.ToStatement(db).Exec(); err != nil {
		return fmt.Errorf("Couldn't delete a subscription: %v", err)
	}
	return nil
}
