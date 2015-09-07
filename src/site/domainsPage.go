// domainsPage.go serves the domains template.
package site

import (
	"database/sql"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// domainsTmplData stores the data that we pass to the template to render the page
type domainsTmplData struct {
	commonPageData
}

// domainsPage serves the recent pages page.
var domainsPage = newPageWithOptions(
	"/domains/",
	domainsRenderer,
	append(baseTmpls,
		"tmpl/domainsPage.tmpl", "tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"),
	newPageOptions{AdminOnly: true})

// domainsRenderer renders the domains page.
func domainsRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	data, err := domainsInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("domains_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("domains_page_served_success")
	return pages.StatusOK(data)
}

// domainsInternalRenderer renders the domains page.
func domainsInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*domainsTmplData, error) {
	var err error
	var data domainsTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the domains
	data.PageMap = make(map[int64]*core.Page)
	data.GroupMap = make(map[int64]*core.Group)
	query := fmt.Sprintf(`
		SELECT id,alias,name,rootPageId,createdAt
		FROM groups
		WHERE isDomain`)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var g core.Group
		err := rows.Scan(
			&g.Id,
			&g.Alias,
			&g.Name,
			&g.RootPageId,
			&g.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan a group: %v", err)
		}
		data.GroupMap[g.Id] = &g
		data.PageMap[g.RootPageId] = &core.Page{PageId: g.RootPageId}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while loading group members: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	return &data, nil
}
