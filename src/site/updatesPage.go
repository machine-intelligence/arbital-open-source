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

type updatedPage struct {
	page

	UpdatedAt string
	Counts    map[string]int // type -> count
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	User         *user.User
	UpdatedPages []*updatedPage
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
	data.UpdatedPages = make([]*updatedPage, 0)
	pageMap := make(map[int64]*updatedPage)
	query := fmt.Sprintf(`
		SELECT p.pageId,p.privacyKey,p.title,u.updatedAt,u.type,u.count
		FROM updates AS u
		LEFT JOIN (
			SELECT * FROM (
				SELECT pageId,privacyKey,title
				FROM pages
				ORDER BY id DESC
			) AS t
			GROUP BY pageId
		) AS p
		ON u.pageId=p.pageId
		WHERE u.userId=%d AND u.seen=0
		ORDER BY u.updatedAt DESC
		LIMIT 50`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var uc updatedPage
		var updateType string
		var count int
		err := rows.Scan(
			&uc.PageId,
			&uc.PrivacyKey,
			&uc.Title,
			&uc.UpdatedAt,
			&updateType,
			&count)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		curPage, ok := pageMap[uc.PageId]
		if !ok {
			uc.Counts = make(map[string]int)
			curPage = &uc
			pageMap[curPage.PageId] = curPage
			data.UpdatedPages = append(data.UpdatedPages, curPage)
		}
		curPage.Counts[updateType] += count
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
		"GetPageUrl": func(p *updatedPage) string {
			privacyAddon := ""
			if p.PrivacyKey.Valid {
				privacyAddon = fmt.Sprintf("/%d", p.PrivacyKey.Int64)
			}
			return fmt.Sprintf("/pages/%d%s", p.PageId, privacyAddon)
		},
	}
	c.Inc("updates_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
