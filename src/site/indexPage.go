// index.go serves the index page.
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	indexPanelLimit = 10
)

// indexTmplData stores the data that we pass to the index.tmpl to render the page
type indexTmplData struct {
	commonPageData
	User                  *user.User
	UserMap               map[int64]*dbUser
	PageMap               map[int64]*page
	RecentlyEditedByMeIds []string
	RecentlyCreatedIds    []string
	MostLikedIds          []string
	RecentlyEditedIds     []string
	MostControversialIds  []string
}

// indexPage serves the index page.
var indexPage = newPage(
	"/",
	indexRenderer,
	append(baseTmpls,
		"tmpl/index.tmpl",
		"tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl"))

// indexRenderer renders the index page.
func indexRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data indexTmplData
	data.User = u
	c := sessions.NewContext(r)
	data.PageMap = make(map[int64]*page)

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
		c.Errorf("error while loading recently edited by me page ids: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load number of red links for recently edited pages.
	err = loadLinks(c, data.PageMap, true)
	if err != nil {
		c.Errorf("error while loading links: %v", err)
		return pages.InternalErrorWith(err)
	}
	for _, p := range data.PageMap {
		p.RedLinkCount = 0
		for _, isPublished := range p.Links {
			if !isPublished {
				p.RedLinkCount++
			}
		}
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
		c.Errorf("error while loading recently created page ids: %v", err)
		return pages.InternalErrorWith(err)
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
		c.Errorf("error while loading most liked page ids: %v", err)
		return pages.InternalErrorWith(err)
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
		c.Errorf("error while loading recently edited page ids: %v", err)
		return pages.InternalErrorWith(err)
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
		c.Errorf("error while loading most controversial page ids: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load pages.
	err = loadPages(c, data.PageMap, u.Id, loadPageOptions{allowUnpublished: true})
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

	c.Inc("index_page_served_success")
	return pages.StatusOK(data)
}
