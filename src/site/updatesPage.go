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

type updatedClaim struct {
	Claim     claim
	UpdatedAt string
	Counts    map[string]int // type -> count
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	User          *user.User
	UpdatedClaims []*updatedClaim
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
	data.UpdatedClaims = make([]*updatedClaim, 0)
	claimMap := make(map[int64]*updatedClaim)
	query := fmt.Sprintf(`
		SELECT c.id,c.privacyKey,c.summary,u.updatedAt,u.type,u.count
		FROM updates AS u
		JOIN claims AS c
		ON u.claimId=c.Id
		WHERE u.userId=%d AND u.seen=0
		ORDER BY u.updatedAt DESC
		LIMIT 50`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var uc updatedClaim
		var updateType string
		var count int
		err := rows.Scan(
			&uc.Claim.Id,
			&uc.Claim.PrivacyKey,
			&uc.Claim.Summary,
			&uc.UpdatedAt,
			&updateType,
			&count)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		curClaim, ok := claimMap[uc.Claim.Id]
		if !ok {
			uc.Counts = make(map[string]int)
			curClaim = &uc
			claimMap[curClaim.Claim.Id] = curClaim
			data.UpdatedClaims = append(data.UpdatedClaims, curClaim)
		}
		curClaim.Counts[updateType] += count
		return nil
	})
	if err != nil {
		c.Errorf("error while loading updates: %v", err)
		return pages.InternalErrorWith(err)
	}

	// TODO: sort Updates

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
