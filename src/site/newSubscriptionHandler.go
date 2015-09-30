// newSubscription.go handles repages for adding a new subscription.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newSubscriptionData contains the data we get in the request.
type newSubscriptionData struct {
	PageId int64 `json:",string"`
	UserId int64 `json:",string"`
}

// newSubscriptionHandler handles requests for adding a new subscription.
func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var emptyData, data newSubscriptionData
	err := decoder.Decode(&data)
	if err != nil || data == emptyData {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// TODO: check if this subscription already exists
	if data.PageId > 0 {
		err = addSubscriptionToPage(db, u.Id, data.PageId)
	} else if data.UserId > 0 {
		err = addSubscriptionToUser(db, u.Id, data.UserId)
	}
	if err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("Couldn't create new subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
