// newGroupHadler.go creates a new group.
package site

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// newGroupData contains data given to us in the request.
type newGroupData struct {
	Name string

	IsDomain   bool
	Alias      string
	RootPageId int64 `json:",string"`
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

	db, err := database.GetDB(c)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}
	if !u.IsLoggedIn {
		return http.StatusForbidden, fmt.Sprintf("Not logged in")
	}
	if u.Karma < 200 {
		return http.StatusForbidden, fmt.Sprintf("You don't have enough karma")
	}
	if !u.IsAdmin {
		return http.StatusForbidden, fmt.Sprintf("Have to be an admin to create domains or groups")
	}

	// Begin the transaction.
	err = db.Transaction(func(tx *database.Tx) error {
		rand.Seed(time.Now().UnixNano())
		groupId := rand.Int63()

		// Create the new group.
		hashmap := make(map[string]interface{})
		hashmap["id"] = groupId
		hashmap["name"] = data.Name
		hashmap["createdAt"] = database.Now()
		if data.IsDomain {
			hashmap["isDomain"] = true
			hashmap["alias"] = data.Alias
			hashmap["rootPageId"] = data.RootPageId
		}
		statement := tx.NewInsertTxStatement("groups", hashmap)
		if _, err = statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't create a group: %v", err)
		}

		if !data.IsDomain {
			// Add the user to the group as an admin.
			hashmap = make(map[string]interface{})
			hashmap["userId"] = u.Id
			hashmap["groupId"] = groupId
			hashmap["canAddMembers"] = true
			hashmap["canAdmin"] = true
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("groupMembers", hashmap)
			if _, err = statement.Exec(); err != nil {
				return fmt.Errorf("Couldn't add a user to a group: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Error commit a transaction: %v\n", err)
	}

	if data.IsDomain {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.RootPageId
		if err := task.IsValid(); err != nil {
			c.Errorf("Invalid task created: %v", err)
		}
		if err := tasks.Enqueue(c, task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}

	return 0, ""
}
