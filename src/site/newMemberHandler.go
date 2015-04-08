// newMemberHadler.go adds a new member for a group.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newMemberData contains data given to us in the request.
type newMemberData struct {
	GroupName string
	UserId    int64 `json:",string"`
}

// newMemberHandler handles requests to add a new member to a group.
func newMemberHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newMemberData
	err := decoder.Decode(&data)
	if err != nil {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if data.GroupName == "" || data.UserId <= 0 {
		c.Errorf("GroupName and UserId have to be set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_member_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		c.Errorf("Not logged in")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check to see if this user can add members.
	var blank int64
	var found bool
	query := fmt.Sprintf(`
		SELECT 1
		FROM groupMembers
		WHERE userId=%d AND groupName="%s" AND canAddMembers`,
		u.Id, data.GroupName)
	found, err = database.QueryRowSql(c, query, &blank)
	if err != nil {
		c.Inc("new_member_fail")
		c.Errorf("Couldn't check for a group member: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !found {
		c.Errorf("You don't have the permission to add a user")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check to see if the proposed member exists.
	query = fmt.Sprintf(`
		SELECT 1
		FROM users
		WHERE id=%d`, data.UserId)
	found, err = database.QueryRowSql(c, query, &blank)
	if err != nil {
		c.Inc("new_member_fail")
		c.Errorf("Couldn't check for a user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !found {
		c.Errorf("New member id doesn't correspond to a user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashmap := make(map[string]interface{})
	hashmap["userId"] = data.UserId
	hashmap["groupName"] = data.GroupName
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("groupMembers", hashmap)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_member_fail")
		c.Errorf("Couldn't add a member: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
