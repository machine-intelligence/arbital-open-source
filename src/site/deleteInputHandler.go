// deleteInputHandler.go handles requests for deleting an input.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// deleteInputData is the object that's put into the daemon queue.
type deleteInputData struct {
	ParentClaimId int64 `json:",string"`
	ChildClaimId  int64 `json:",string"`
}

// deleteInputHandler handles requests for deleting an input.
func deleteInputHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data deleteInputData
	err := decoder.Decode(&data)
	if err != nil || data.ParentClaimId == 0 || data.ChildClaimId == 0 {
		c.Inc("delete_input_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("delete_input_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !u.IsLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	query := fmt.Sprintf(`
		DELETE FROM inputs
		WHERE parentId=%d AND childId=%d`, data.ParentClaimId, data.ChildClaimId)
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("delete_input_fail")
		c.Errorf("Couldn't delete an input: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
