// domainsPage.go serves the domains template.
package site

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
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
		"tmpl/domainsPage.tmpl",
		"tmpl/angular.tmpl.js", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"),
	pages.PageOptions{AdminOnly: true})

// domainsRenderer renders the domains page.
func domainsRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var err error
	var data domainsTmplData
	data.User = u

	// Load the domains
	data.PageMap = make(map[int64]*core.Page)
	data.GroupMap = make(map[int64]*core.Group)
	rows := db.NewStatement(`
		SELECT id,alias,name,rootPageId,createdAt
		FROM groups
		WHERE isDomain`).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
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
		return pages.Fail("error while loading group members", err)
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, u.Id, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	return pages.StatusOK(&data)
}
