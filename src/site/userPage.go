// userPage.go serves the user template.
package site

import (
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
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
		"tmpl/userPage.tmpl",
		"tmpl/angular.tmpl.js"),
	pages.PageOptions{})

// userRenderer renders the user page.
func userRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var err error
	var data userTmplData
	data.User = u
	data.PageMap = make(map[int64]*core.Page)
	data.EditMap = make(map[int64]*core.Page)
	data.UserMap = make(map[int64]*core.User)

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

	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

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
	data.RecentlyCreatedIds, err = core.LoadPageIds(rows, data.PageMap, pageOptions)
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
	data.RecentlyEditedIds, err = core.LoadPageIds(rows, data.PageMap, pageOptions)
	if err != nil {
		return pages.Fail("error while loading recently edited page ids", err)
	}

	if data.User.Id == data.AuthorId {
		// Load pages with unpublished drafts
		rows = db.NewStatement(`
			SELECT p.pageId,p.title,p.createdAt,i.currentEdit>0
			FROM pages AS p
			JOIN pageInfos AS i
			ON (p.pageId = i.pageId)
			WHERE p.creatorId=? AND p.type!=? AND p.edit>i.currentEdit
			GROUP BY p.pageId
			ORDER BY p.createdAt DESC
			LIMIT ?`).Query(data.AuthorId, core.CommentPageType, indexPanelLimit)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId int64
			var title, createdAt string
			var wasPublished bool
			err := rows.Scan(&pageId, &title, &createdAt, &wasPublished)
			if err != nil {
				return fmt.Errorf("failed to scan: %v", err)
			}
			data.PagesWithDraftIds = append(data.PagesWithDraftIds, fmt.Sprintf("%d", pageId))
			page := core.AddPageIdToMap(pageId, data.EditMap)
			if title == "" {
				title = "*Untitled*"
			}
			page.Title = title
			page.CreatedAt = createdAt
			page.WasPublished = wasPublished
			return nil
		})
		if err != nil {
			return pages.Fail("error while loading pages with drafts ids", err)
		}

		// Load page ids with the most todos.
		rows = db.NewStatement(`
			SELECT l.parentId
			FROM (
				SELECT l.parentId AS parentId,l.childAlias AS childAlias,p.todoCount AS parentTodoCount
				FROM links AS l
				JOIN pages AS p
				ON (l.parentId=p.pageId)
				WHERE p.creatorId=? AND p.type!="comment" AND p.isCurrentEdit
			) AS l
			LEFT JOIN pages AS p
			ON (l.childAlias=p.alias OR l.childAlias=p.pageId)
			GROUP BY 1
			ORDER BY (SUM(ISNULL(p.pageId)) + max(l.parentTodoCount)) DESC
			LIMIT ?`).Query(data.AuthorId, indexPanelLimit)
		data.MostTodosIds, err = core.LoadPageIds(rows, data.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading most todos page ids", err)
		}
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
	data.RecentlyEditedCommentIds, err = core.LoadPageIds(rows, data.PageMap, core.TitlePlusLoadOptions)
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
		data.RecentlyVisitedIds, err = core.LoadPageIds(rows, data.PageMap, pageOptions)
		if err != nil {
			return pages.Fail("error while loading recently visited page ids", err)
		}
	}

	// Load pages.
	data.UserMap[data.AuthorId] = &core.User{Id: data.AuthorId}
	err = core.ExecuteLoadPipeline(db, u, data.PageMap, data.UserMap, data.MasteryMap)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.StatusOK(&data)
}
