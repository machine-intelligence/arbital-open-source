// domainsPage.go serves the domains template.
package site

import (
	"zanaduu3/src/core"
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
		"tmpl/angular.tmpl.js"),
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
	err = core.LoadDomainIds(db, u, nil, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading group members", err)
	}

	// Load pages.
	core.AddUserGroupIdsToPageMap(data.User, data.PageMap)
	err = core.LoadPages(db, data.PageMap, u, nil)
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	return pages.StatusOK(&data)
}
