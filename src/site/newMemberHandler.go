// newMemberHadler.go adds a new member for a group.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newMemberData contains data given to us in the request.
type newMemberData struct {
	GroupId int64 `json:",string"`
	UserId  int64 `json:",string"`
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
	if data.GroupId <= 0 || data.UserId <= 0 {
		c.Errorf("GroupId and UserId have to be set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB(c)
	if err != nil {
		c.Inc("new_member_fail")
		c.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r, db)
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
	row := db.NewStatement(`
		SELECT 1
		FROM groupMembers
		WHERE userId=%d AND groupId=%d AND canAddMembers
		`).QueryRow(u.Id, data.GroupId)
	found, err = row.Scan(&blank)
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
	row = db.NewStatement(`
		SELECT 1
		FROM users
		WHERE id=%d`).QueryRow(data.UserId)
	found, err = row.Scan(&blank)
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
	hashmap["groupId"] = data.GroupId
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("groupMembers", hashmap)
	if _, err = statement.Exec(); err != nil {
		c.Inc("new_member_fail")
		c.Errorf("Couldn't add a member: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
