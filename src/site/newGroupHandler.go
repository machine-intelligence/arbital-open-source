// newGroupHadler.go creates a new group.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newGroupData contains data given to us in the request.
type newGroupData struct {
	Name string
}

// newGroupHandler handles requests to add a new group to a group.
func newGroupHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	header, str := newGroupProcessor(w, r)
	if header > 0 {
		if header == http.StatusInternalServerError {
			c.Inc(strings.Trim(r.URL.Path, "/") + "Fail")
		}
		c.Errorf("%s", str)
		w.WriteHeader(header)
	}
	if len(str) > 0 {
		fmt.Fprintf(w, "%s", str)
	}
}

func newGroupProcessor(w http.ResponseWriter, r *http.Request) (int, string) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newGroupData
	err := decoder.Decode(&data)
	if err != nil {
		return http.StatusBadRequest, fmt.Sprintf("Couldn't decode json: %v", err)
	}
	if data.Name == "" {
		return http.StatusBadRequest, fmt.Sprintf("Name has to be set")
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}
	if !u.IsLoggedIn {
		return http.StatusForbidden, fmt.Sprintf("Not logged in")
	}
	if u.Karma <= 200 {
		return http.StatusForbidden, fmt.Sprintf("You don't have enough karma")
	}

	// Begin the transaction.
	tx, err := database.NewTransaction(c)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("failed to create a transaction: %v\n", err)
	}

	// Create the new group.
	hashmap := make(map[string]interface{})
	hashmap["name"] = data.Name
	hashmap["createdAt"] = database.Now()
	query := database.GetInsertSql("groups", hashmap)
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't create a group: %v", err)
	}

	// Add the user to the group as an admin.
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["groupName"] = data.Name
	hashmap["canAddMembers"] = true
	hashmap["canAdmin"] = true
	hashmap["createdAt"] = database.Now()
	query = database.GetInsertSql("groupMembers", hashmap)
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't add a user to a group: %v", err)
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, fmt.Sprintf("Error commit a transaction: %v\n", err)
	}

	return 0, ""
}
