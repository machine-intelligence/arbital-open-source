// updatesPage.go serves the update page.
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

type update struct {
	Count         int
	FromCommentId int64
	FromUser      *dbUser
	FromTag       *tag
}

type updatedPage struct {
	page

	UpdatedAt string
	Updates   map[string]*update // update type -> update
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	User         *user.User
	UpdatedPages []*updatedPage
}

// updatesPage serves the updates page.
var updatesPage = pages.Add(
	"/updates/",
	updatesRenderer,
	append(baseTmpls,
		"tmpl/updatesPage.tmpl", "tmpl/tag.tmpl", "tmpl/userName.tmpl", "tmpl/navbar.tmpl")...)

// updatesRenderer renders the updates page.
func updatesRenderer(w http.ResponseWriter, r *http.Request) *pages.Result {
	var data updatesTmplData
	c := sessions.NewContext(r)

	// Load user, if possible
	var err error
	data.User, err = user.LoadUserFromDb(c)
	if err != nil {
		c.Errorf("Couldn't load user: %v", err)
		return pages.InternalErrorWith(err)
	}
	if !data.User.IsLoggedIn {
		return pages.UnauthorizedWith(fmt.Errorf("Not logged in"))
	}

	// Load the updates
	data.UpdatedPages = make([]*updatedPage, 0)
	pageMap := make(map[int64]*updatedPage)
	userMap := make(map[int64]*dbUser)
	tagMap := make(map[int64]*tag)
	query := fmt.Sprintf(`
		SELECT p.pageId,p.privacyKey,p.title,u.updatedAt,u.type,u.count,u.fromCommentId,u.fromUserId,u.fromTagId
		FROM updates AS u
		LEFT JOIN (
			SELECT * FROM (
				SELECT pageId,privacyKey,title
				FROM pages
				WHERE deletedBy=0
				ORDER BY id DESC
			) AS t
			GROUP BY pageId
		) AS p
		ON u.contextPageId=p.pageId
		WHERE u.userId=%d AND u.seen=0
		ORDER BY u.updatedAt DESC
		LIMIT 50`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var uc updatedPage
		var u update
		var updateType string
		var fromUserId, fromTagId int64
		err := rows.Scan(
			&uc.PageId,
			&uc.PrivacyKey,
			&uc.Title,
			&uc.UpdatedAt,
			&updateType,
			&u.Count,
			&u.FromCommentId,
			&fromUserId,
			&fromTagId)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		// Create/get the current page.
		curPage, ok := pageMap[uc.PageId]
		if !ok {
			uc.Updates = make(map[string]*update)
			curPage = &uc
			pageMap[curPage.PageId] = curPage
			data.UpdatedPages = append(data.UpdatedPages, curPage)
		}

		// Create/get the current update.
		curUpdate, ok2 := curPage.Updates[updateType]
		if !ok2 {
			curUpdate = &u
			curPage.Updates[updateType] = curUpdate
		} else {
			curUpdate.Count += u.Count
		}

		// If there is a user, proces it.
		if fromUserId > 0 {
			curUser, ok3 := userMap[fromUserId]
			if !ok3 {
				curUser = &dbUser{Id: fromUserId}
				userMap[fromUserId] = curUser
			}
			curUpdate.FromUser = curUser
		}

		// If there is a tag, proces it.
		if fromTagId > 0 {
			curTag, ok3 := tagMap[fromTagId]
			if !ok3 {
				curTag = &tag{Id: fromTagId}
				tagMap[fromTagId] = curTag
			}
			curUpdate.FromTag = curTag
		}
		return nil
	})
	if err != nil {
		c.Errorf("error while loading updates: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the names for all users.
	err = loadUsersInfo(c, userMap)
	if err != nil {
		c.Errorf("error while loading user names: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load the names for all tags.
	err = loadTagNames(c, tagMap)
	if err != nil {
		c.Errorf("error while loading tag names: %v", err)
		return pages.InternalErrorWith(err)
	}

	// TODO: sort Updates?

	funcMap := template.FuncMap{
		"UserId":     func() int64 { return data.User.Id },
		"IsAdmin":    func() bool { return data.User.IsAdmin },
		"IsLoggedIn": func() bool { return data.User.IsLoggedIn },
		"GetPageUrl": func(p *updatedPage) string {
			return getPageUrl(&p.page)
		},
		"GetUserUrl": func(userId int64) string {
			return getUserUrl(userId)
		},
		"GetTagUrl": func(tagId int64) string {
			return getTagUrl(tagId)
		},
	}
	c.Inc("updates_page_served_success")
	return pages.StatusOK(data).SetFuncMap(funcMap)
}
