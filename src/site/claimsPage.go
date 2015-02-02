// claimsPage.go serves the claim page.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// claimsTmplData stores the data that we pass to the claims.tmpl to render the page
type claimsTmplData struct {
	User   *user.User
	Claims []*claim
}

// claimsPage serves the claims page.
var claimsPage = pages.Add(
	"/claims/all/",
	claimsRenderer,
	append(baseTmpls,
		"tmpl/claims.tmpl", "tmpl/navbar.tmpl")...)

// claimsRenderer renders the claim page.
func claimsRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data claimsTmplData
	c := sessions.NewContext(r)

	// Load the claims
	data.Claims = make([]*claim, 0, 50)
	query := fmt.Sprintf(`
		SELECT id,text
		FROM claims
		WHERE privacyKey IS NULL
		ORDER BY id DESC
		LIMIT 50`)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var q claim
		err := rows.Scan(
			&q.Id,
			&q.Text)
		if err != nil {
			return fmt.Errorf("failed to scan a claim: %v", err)
		}

		// Load tags.
		err = loadTags(c, &q)
		if err != nil {
			return fmt.Errorf("Couldn't retrieve claim tags: %v", err)
		}

		data.Claims = append(data.Claims, &q)
		return nil
	})
	if err != nil {
		c.Errorf("error while loading claims: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load user, if possible
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
	}
	c.Inc("claims_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
