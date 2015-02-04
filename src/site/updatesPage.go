// updatesPage.go serves the update page.
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

type update struct {
	Id        int64
	Claim     claim
	CommentId int64
	Type      string
	CreatedAt string
	UpdatedAt string
	Count     int
	Seen      bool
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	User    *user.User
	Updates []*update
}

// updatesPage serves the updates page.
var updatesPage = pages.Add(
	"/updates/",
	updatesRenderer,
	append(baseTmpls,
		"tmpl/updatesPage.tmpl", "tmpl/navbar.tmpl")...)

// updatesRenderer renders the updates page.
func updatesRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data updatesTmplData
	c := sessions.NewContext(r)

	// Load user, if possible
	var err error
	data.User, err = user.LoadUser(w, r)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}
	if !data.User.IsLoggedIn {
		return pages.UnauthorizedWith(fmt.Errorf("Not logged in"))
	}

	// Load the updates
	data.Updates = make([]*update, 0)
	query := fmt.Sprintf(`
		SELECT u.id,c.id,c.privacyKey,u.commentId,u.type,u.createdAt,u.updatedAt,u.count,u.seen
		FROM updates AS u
		JOIN claims AS c
		ON u.claimId=c.Id
		WHERE u.userId=%d
		ORDER BY u.updatedAt DESC
		LIMIT 50`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var u update
		err := rows.Scan(
			&u.Id,
			&u.Claim.Id,
			&u.Claim.PrivacyKey,
			&u.CommentId,
			&u.Type,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Count,
			&u.Seen)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		data.Updates = append(data.Updates, &u)
		return nil
	})
	if err != nil {
		c.Errorf("error while loading updates: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
		"GetClaimUrl": func(c *claim) string {
			privacyAddon := ""
			if c.PrivacyKey.Valid {
				privacyAddon = fmt.Sprintf("/%d", c.PrivacyKey.Int64)
			}
			return fmt.Sprintf("/claims/%d%s", c.Id, privacyAddon)
		},
	}
	c.Inc("updates_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
