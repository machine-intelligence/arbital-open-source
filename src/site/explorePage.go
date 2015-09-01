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
)

// exploreTmplData stores the data that we pass to the template to render the page
type exploreTmplData struct {
	commonPageData
}

// explorePage serves the Explore page.
var explorePage = newPage(
	"/explore/",
	exploreRenderer,
	append(baseTmpls,
		"tmpl/explorePage.tmpl", "tmpl/angular.tmpl.js", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// exploreRenderer renders the page page.
func exploreRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data exploreTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the pages
	data.PageMap = make(map[int64]*core.Page)
	query := fmt.Sprintf(`
		SELECT parentPair.parentId
		FROM pagePairs AS parentPair
		LEFT JOIN pagePairs AS grandParentPair
		ON (parentPair.parentId=grandParentPair.childId)
		WHERE grandParentPair.parentId IS NULL
		GROUP BY 1
		LIMIT 50`)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
		c.Errorf("error while loading page pairs: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the children
	err = loadChildrenIds(c, data.PageMap, loadChildrenIdsOptions{LoadHasChildren: true})
	if err != nil {
		c.Errorf("error while loading children: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, core.LoadPageOptions{})
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		return pages.InternalErrorWith(err)
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
		c.Errorf("Couldn't load aux data: %v", err)
		return pages.InternalErrorWith(err)
	}

	c.Inc("explore_page_served_success")
	return pages.StatusOK(data)
}
