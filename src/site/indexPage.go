// index.go serves the index page.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

const (
	indexPanelLimit = 10
)

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	commonPageData
	RecentlyEditedByMeIds []string
	RecentlyVisitedIds    []string
	RecentlyCreatedIds    []string
	MostLikedIds          []string
	RecentlyEditedIds     []string
	MostControversialIds  []string
}

// indexPage serves the index page.
var indexPage = newPageWithOptions(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/index.tmpl",
		"tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl"),
	newPageOptions{})

// domainIndexPage serves the index page for a domain.
var domainIndexPage = newPageWithOptions(
	"/{alias:[A-Za-z0-9_-]+}",
	indexRenderer,
	append(baseTmpls,
		"tmpl/index.tmpl",
		"tmpl/pageHelpers.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl"),
	newPageOptions{})

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	data, err := indexInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("index_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}

// indexInternalRenderer renders the index page.
func indexInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*indexTmplData, error) {
	var err error
	var data indexTmplData
	data.User = u
	c := sessions.NewContext(r)
	data.PageMap = make(map[int64]*core.Page)

	// Check if the user is looking at a specific domain
	domainAlias := mux.Vars(r)["domain1"]
	//domainConstraint := ""
	if mux.Vars(r)["subdomaindot"] != "" {
		c.Debugf("========================%s", domainAlias)
	}

	// Load recently edited by me page ids.
	query := fmt.Sprintf(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE creatorId=%d
			GROUP BY pageId
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.User.Id, indexPanelLimit)
	data.RecentlyEditedByMeIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently edited by me page ids: %v", err)
	}

	// Load number of red links for recently edited pages.
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

	// Load recently visited page ids.
	query = fmt.Sprintf(`
		SELECT v.pageId
		FROM (
			SELECT pageId,max(createdAt) AS createdAt
			FROM visits
			WHERE userId=%d
			GROUP BY 1
		) AS v
		ORDER BY v.createdAt DESC
		LIMIT %d`, data.User.Id, indexPanelLimit)
	data.RecentlyVisitedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently visited page ids: %v", err)
	}

	// Load recently created page ids.
	query = fmt.Sprintf(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE NOT isSnapshot AND NOT isAutosave 
			ORDER BY createdAt DESC
		) AS p
		GROUP BY p.pageId
		ORDER BY MIN(createdAt) DESC
		LIMIT %d`, indexPanelLimit)
	data.RecentlyCreatedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently created page ids: %v", err)
	}

	// Load most liked page ids.
	query = fmt.Sprintf(`
		SELECT pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM likes
				ORDER BY id DESC
			) AS l1
			GROUP BY userId,pageId
		) AS l2
		GROUP BY pageId
		ORDER BY SUM(value) DESC
		LIMIT %d`, indexPanelLimit)
	data.MostLikedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading most liked page ids: %v", err)
	}

	// Load recently edited page ids.
	query = fmt.Sprintf(`
		SELECT p.pageId
		FROM (
			SELECT pageId,max(createdAt) AS createdAt
			FROM pages
			WHERE NOT isSnapshot AND NOT isAutosave 
			GROUP BY pageId
			HAVING(SUM(1) > 1)
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT %d`, indexPanelLimit)
	data.RecentlyEditedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently edited page ids: %v", err)
	}

	// Load most controversial page ids.
	// TODO: make sure the page still has voting turned on
	query = fmt.Sprintf(`
		SELECT pageId
		FROM (
			SELECT *
			FROM (
				SELECT *
				FROM votes
				ORDER BY id DESC
			) AS v1
			GROUP BY userId,pageId
		) AS v2
		GROUP BY pageId
		ORDER BY VAR_POP(value) DESC
		LIMIT %d`, indexPanelLimit)
	data.MostControversialIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading most controversial page ids: %v", err)
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, core.LoadPageOptions{AllowUnpublished: true})
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(c, u.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load aux data: %v", err)
	}

	return &data, nil
}
