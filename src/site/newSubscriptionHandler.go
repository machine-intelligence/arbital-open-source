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
	PageId    int64 `json:",string"`
	CommentId int64 `json:",string"`
	UserId    int64 `json:",string"`
	TagId     int64 `json:",string"`
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

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
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

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["createdAt"] = database.Now()
	if data.PageId > 0 {
		hashmap["toPageId"] = data.PageId
	} else if data.CommentId > 0 {
		hashmap["toCommentId"] = data.CommentId
	} else if data.UserId > 0 {
		hashmap["toUserId"] = data.UserId
	} else if data.TagId > 0 {
		hashmap["toTagId"] = data.TagId
	}
	sql := database.GetInsertSql("subscriptions", hashmap)
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("Couldn't create new subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
