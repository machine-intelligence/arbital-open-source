// userPage.go serves the user template.
package site

import (
	"fmt"
	"net/http"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"

	"github.com/gorilla/mux"
)

const (
	historyLimit = 100
)

// userTmplData stores the data that we pass to the template to render the page
type userTmplData struct {
	commonPageData

	// Id of the user we are looking at
	AuthorId int64
	// True iff the current user is subsribed to the author
	IsSubscribedToUser       bool
	RecentlyEditedIds        []string
	MostTodosIds             []string
	RecentlyEditedCommentIds []string
	RecentlyVisitedIds       []string
}

// userPage serves the recent pages page.
var userPage = newPageWithOptions(
	"/user/{authorId:[0-9]+}",
	userRenderer,
	append(baseTmpls,
		"tmpl/userPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl",
		"tmpl/footer.tmpl", "tmpl/angular.tmpl.js"),
	newPageOptions{LoadUserGroups: true})

// userRenderer renders the user page.
func userRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	c := sessions.NewContext(r)
	data, err := userInternalRenderer(w, r, u)
	if err != nil {
		c.Inc("user_page_served_fail")
		c.Errorf("%s", err)
		return showError(w, r, fmt.Errorf("%s", err))
	}
	c.Inc("user_page_served_success")
	return pages.StatusOK(data)
}

// userInternalRenderer renders the page page.
func userInternalRenderer(w http.ResponseWriter, r *http.Request, u *user.User) (*userTmplData, error) {
	var err error
	var data userTmplData
	data.User = u
	c := sessions.NewContext(r)
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	data.GroupMap = make(map[int64]*core.Group)

	// Check parameter limiting the user/creator of the pages
	data.AuthorId, err = strconv.ParseInt(mux.Vars(r)["authorId"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse authorId: %s", mux.Vars(r)["authorId"])
	}

	var throwaway int
	query := fmt.Sprintf(`
		SELECT 1
		FROM subscriptions
		WHERE userId=%d AND toUserId=%d`, data.User.Id, data.AuthorId)
	data.IsSubscribedToUser, err = database.QueryRowSql(c, query, &throwaway)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve subscription: %v", err)
	}

	// Load recently edited by me page ids.
	query = fmt.Sprintf(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE creatorId=%d AND type!="comment"
			GROUP BY pageId
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.AuthorId, indexPanelLimit)
	data.RecentlyEditedIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently edited page ids: %v", err)
	}

	if data.User.Id == data.AuthorId {
		// Load page ids with the most todos.
		query = fmt.Sprintf(`
			SELECT l.parentId
			FROM (
				SELECT l.parentId AS parentId,l.childAlias AS childAlias
				FROM links AS l
				JOIN pages AS p
				ON (l.parentId=p.pageId)
				WHERE p.creatorId=%d AND p.type!="comment" AND p.isCurrentEdit
			) AS l
			LEFT JOIN pages AS p
			ON (l.childAlias=p.alias)
			GROUP BY 1
			ORDER BY SUM(IF(p.pageId IS NULL, 1, 0)) DESC
			LIMIT %d`, data.AuthorId, indexPanelLimit)
		data.MostTodosIds, err = loadPageIds(c, query, data.PageMap)
		if err != nil {
			return nil, fmt.Errorf("error while loading most todos page ids: %v", err)
		}
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

	// Load recently edited by me comment ids
	query = fmt.Sprintf(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE creatorId=%d AND type="comment"
			GROUP BY pageId
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.AuthorId, indexPanelLimit)
	data.RecentlyEditedCommentIds, err = loadPageIds(c, query, data.PageMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading recently edited by me page ids: %v", err)
	}

	// Load the following info for yourself only.
	if data.User.Id == data.AuthorId {
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
			LIMIT %d`, data.AuthorId, indexPanelLimit)
		data.RecentlyVisitedIds, err = loadPageIds(c, query, data.PageMap)
		if err != nil {
			return nil, fmt.Errorf("error while loading recently visited page ids: %v", err)
		}
	}

	// Load pages.
	err = core.LoadPages(c, data.PageMap, u.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading pages: %v", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(c, data.User.Id, data.PageMap, nil)
	if err != nil {
		return nil, fmt.Errorf("error while loading aux data: %v", err)
	}

	// Load all the groups.
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load group names: %v", err)
	}

	// Load all the users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	data.UserMap[data.AuthorId] = &core.User{Id: data.AuthorId}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, data.UserMap)
	if err != nil {
		return nil, fmt.Errorf("error while loading users: %v", err)
	}

	return &data, nil
}
