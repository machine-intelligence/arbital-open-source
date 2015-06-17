// explorePage.go serves the explore template.
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

// exploreTmplData stores the data that we pass to the template to render the page
type exploreTmplData struct {
	User    *user.User
	UserMap map[int64]*dbUser
	PageMap map[int64]*page
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
	pageIds := make([]string, 0, 50)
	data.PageMap = make(map[int64]*page)
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
		p := &page{PageId: pageId}
		data.PageMap[pageId] = p
		pageIds = append(pageIds, fmt.Sprintf("%d", pageId))
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
	err = loadPages(c, data.PageMap, u.Id, loadPageOptions{})
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load likes.
	err = loadLikes(c, data.User.Id, data.PageMap)
	if err != nil {
		c.Errorf("Couldn't retrieve page likes: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{}

	c.Inc("pages_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
