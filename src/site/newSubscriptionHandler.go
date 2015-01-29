// newSubscription.go handles requestions for adding a new subscription.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newSubscriptionData is the object that's put into the daemon queue.
type newSubscriptionData struct {
	QuestionId int64 `json:",string"`
	CommentId  int64 `json:",string"`
}

// newSubscriptionHandler handles requests for adding a new subscription.
func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newSubscriptionData
	err := decoder.Decode(&data)
	if err != nil {
		c.Inc("new_subscription_fail")
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

	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["createdAt"] = database.Now()
	if data.QuestionId > 0 {
		hashmap["questionId"] = data.QuestionId
	} else if data.CommentId > 0 {
		hashmap["commentId"] = data.CommentId
	} else {
		c.Inc("new_subscription_fail")
		c.Errorf("At least one id has to be given: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sql := database.GetInsertSql("subscriptions", hashmap)
	if _, err = database.ExecuteSql(c, sql); err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("Couldn't create new subscription: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
