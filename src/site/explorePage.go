// explorePage.go serves the explore template.
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

	"github.com/gorilla/mux"
)

// exploreTmplData stores the data that we pass to the template to render the page
type exploreTmplData struct {
	commonPageData
}

// explorePage serves the Explore page.
var exploreAllPage = newPage(
	"/explore/",
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// explorePage serves the Explore page.
var explorePage = newPage(
	"/explore/{domain:[A-Za-z0-9_-]+}",
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// exploreRenderer renders the explore page.
func exploreRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)

	data, err := exploreInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("explore_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("explore_page_served_success")
	return pages.StatusOK(data)
}

// exploreInternalRenderer renders the page page.
func exploreInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*exploreTmplData, error) {
	var err error
	var data exploreTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the domain.
	domainAlias := mux.Vars(r)["domain"]
	if domainAlias != "" {
		data.Domain = &core.Group{Alias: domainAlias}
		data.User.DomainAlias = data.Domain.Alias
		query := fmt.Sprintf(`
			SELECT id,name,rootPageId
			FROM groups
			WHERE alias='%s'`, data.Domain.Alias)
		foundDomain, err := database.QueryRowSql(c, query, &data.Domain.Id, &data.Domain.Name, &data.Domain.RootPageId)
		if err != nil {
			return nil, fmt.Errorf("Couldn't retrieve subscription: %v", err)
		} else if !foundDomain {
			return nil, fmt.Errorf("Couldn't find the domain: %s", data.Domain.Alias)
		}
	}

	// Load the root page(s)
	data.PageMap = make(map[int64]*core.Page)
	if domainAlias == "" {
		query := fmt.Sprintf(`
			SELECT parentPair.parentId
			FROM pagePairs AS parentPair
			LEFT JOIN pagePairs AS grandParentPair
			ON (parentPair.parentId=grandParentPair.childId)
			WHERE grandParentPair.parentId IS NULL
			GROUP BY 1
			LIMIT 50`)
		err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var pageId int64
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("failed to scan a page id: %v", err)
			}
			p := &core.Page{PageId: pageId}
			data.PageMap[pageId] = p
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error while loading page pairs: %v", err)
		}
	} else {
		data.PageMap[data.Domain.RootPageId] = &core.Page{PageId: data.Domain.RootPageId}
	}

	// Load the children
	err = loadChildrenIds(c, data.PageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		return nil, fmt.Errorf("error while loading children: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Filter unpublished pages.
	for id, p := range data.PageMap {
		if !p.IsCurrentEdit {
			delete(data.PageMap, id)
		}
	}

	// Load auxillary data.
	err = loadAuxPageData(c, data.User.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load aux data: %v", err)
	}

	// Load number of red links.
	err = loadLinks(c, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading links: %v", err)
	}
	for _, p := range data.PageMap {
		p.RedLinkCount = 0
		for _, title := range p.Links {
			if title == "" {
				p.RedLinkCount++
			}
		}
	}

	return &data, nil
}
