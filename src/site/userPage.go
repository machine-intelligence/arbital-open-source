// userPage.go serves the user template.
package site

import (
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"

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
	RecentlyCreatedIds       []string
	RecentlyEditedIds        []string
	PagesWithDraftIds        []string
	MostTodosIds             []string
	RecentlyEditedCommentIds []string
	RecentlyVisitedIds       []string
}

// userPage serves the recent pages page.
var userPage = newPageWithOptions(
	"/user/{authorId:[0-9]+}",
	userRenderer,
	append(baseTmpls,
		"tmpl/userPage.tmpl", "tmpl/navbar.tmpl",
		"tmpl/footer.tmpl", "tmpl/angular.tmpl.js"),
	pages.PageOptions{})

// userRenderer renders the user page.
func userRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var err error
	var data userTmplData
	data.User = u
	data.PageMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)
	data.GroupMap = make(map[int64]*core.Group)

	// Check parameter limiting the user/creator of the pages
	data.AuthorId, err = strconv.ParseInt(mux.Vars(params.R)["authorId"], 10, 64)
	if err != nil {
		return pages.Fail(fmt.Sprintf("Couldn't parse authorId: %s", mux.Vars(params.R)["authorId"]), err)
	}

	var throwaway int
	row := db.NewStatement(`
		SELECT 1
		FROM subscriptions
		WHERE userId=? AND toUserId=?`).QueryRow(data.User.Id, data.AuthorId)
	data.IsSubscribedToUser, err = row.Scan(&throwaway)
	if err != nil {
		return pages.Fail("Couldn't retrieve subscription", err)
	}

	// Load recently created by me page ids.
	rows := db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,edit
			FROM pages
			WHERE creatorId=? AND type!="comment"
		) AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId && p.edit=pi.currentEdit)
		ORDER BY pi.createdAt DESC
		LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
	data.RecentlyCreatedIds, err = loadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading recently created page ids", err)
	}

	// Load recently edited by me page ids.
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,max(edit) AS maxEdit,min(edit) AS minEdit,max(createdAt) AS createdAt
			FROM pages
			WHERE creatorId=? AND type!="comment" AND NOT isAutosave
			GROUP BY 1
		) AS p
		WHERE maxEdit>minEdit
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
	data.RecentlyEditedIds, err = loadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading recently edited page ids", err)
	}

	if data.User.Id == data.AuthorId {
		// Load recently edited by me page ids.
		rows = db.NewStatement(`
			SELECT p.pageId
			FROM (
				SELECT pageId,createdAt
				FROM pages
				WHERE creatorId=? AND type!="comment" AND isAutosave
				GROUP BY pageId
			) AS p
			ORDER BY createdAt DESC
			LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
		data.PagesWithDraftIds, err = loadPageIds(rows, data.PageMap)
		if err != nil {
			return pages.Fail("error while loading pages with drafts ids", err)
		}

		// Load page ids with the most todos.
		rows = db.NewStatement(`
			SELECT l.parentId
			FROM (
				SELECT l.parentId AS parentId,l.childAlias AS childAlias
				FROM links AS l
				JOIN pages AS p
				ON (l.parentId=p.pageId)
				WHERE p.creatorId=? AND p.type!="comment" AND p.isCurrentEdit
			) AS l
			LEFT JOIN pages AS p
			ON (l.childAlias=p.alias)
			GROUP BY 1
			ORDER BY SUM(IF(p.pageId IS NULL, 1, 0)) DESC
			LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
		data.MostTodosIds, err = loadPageIds(rows, data.PageMap)
		if err != nil {
			return pages.Fail("error while loading most todos page ids", err)
		}
	}

	// Load number of red links for recently edited pages.
	err = loadRedLinkCount(db, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading links", err)
	}

	// Load recently edited by me comment ids
	rows = db.NewStatement(`
		SELECT p.pageId
		FROM (
			SELECT pageId,createdAt
			FROM pages
			WHERE creatorId=? AND type="comment"
			GROUP BY pageId
		) AS p
		ORDER BY p.createdAt DESC
		LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
	data.RecentlyEditedCommentIds, err = loadPageIds(rows, data.PageMap)
	if err != nil {
		return pages.Fail("error while loading recently edited by me page ids", err)
	}

	// Load the following info for yourself only.
	if data.User.Id == data.AuthorId {
		// Load recently visited page ids.
		rows = db.NewStatement(`
			SELECT v.pageId
			FROM (
				SELECT pageId,max(createdAt) AS createdAt
				FROM visits
				WHERE userId=?
				GROUP BY 1
			) AS v
			ORDER BY v.createdAt DESC
			LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
		data.RecentlyVisitedIds, err = loadPageIds(rows, data.PageMap)
		if err != nil {
			return pages.Fail("error while loading recently visited page ids", err)
		}
	}

	// Load pages.
	err = core.LoadPages(db, data.PageMap, u.Id, &core.LoadPageOptions{AllowUnpublished: true})
	if err != nil {
		return pages.Fail("error while loading pages", err)
	}

	// Load auxillary data.
	err = loadAuxPageData(db, data.User.Id, data.PageMap, nil)
	if err != nil {
		return pages.Fail("error while loading aux data", err)
	}

	// Load all the groups.
	err = loadGroupNames(db, u, data.GroupMap)
	if err != nil {
		return pages.Fail("Couldn't load group names", err)
	}

	// Load all the users.
	data.UserMap[u.Id] = &core.User{Id: u.Id}
	data.UserMap[data.AuthorId] = &core.User{Id: data.AuthorId}
	for _, p := range data.PageMap {
		data.UserMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(db, data.UserMap)
	if err != nil {
		return pages.Fail("error while loading users", err)
	}

	return pages.StatusOK(&data)
}
