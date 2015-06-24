// aliasesJsonHandler.go contains the handler for returning JSON with all pages'
// aliases and titles.
package site

import (
	"database/sql"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/schema"
)

// aliasesJsonData contains parameters passed in via the request.
type aliasesJsonData struct {
}

// aliasesJsonHandler handles the request.
func aliasesJsonHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	returnData := make(map[string]interface{})

	// Decode data
	var data aliasesJsonData
	r.ParseForm()
	err := schema.NewDecoder().Decode(&data, r.Form)
	if err != nil {
		c.Errorf("Couldn't decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load aliases.
	aliases := make([]*alias, 0)
	query := fmt.Sprintf(`
		SELECT pageId,alias,title
		FROM pages
		WHERE isCurrentEdit AND (groupName="" OR groupName IN (SELECT groupName FROM groupMembers WHERE userId=%d))`,
		u.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var a alias
		err := rows.Scan(&a.PageId, &a.FullName, &a.PageTitle)
		if err != nil {
			return fmt.Errorf("failed to scan for aliases: %v", err)
		}
		aliases = append(aliases, &a)
		return nil
	})
	if err != nil {
		c.Errorf("Couldn't load aliases: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return the pages in JSON format.
	for _, a := range aliases {
		returnData[a.FullName] = a
	}
	err = writeJson(w, returnData)
	if err != nil {
		c.Inc("pages_handler_fail")
		c.Errorf("Couldn't write json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
