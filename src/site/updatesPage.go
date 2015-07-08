// updatesPage.go serves the update page.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

type update struct {
	Count    int
	FromUser *dbUser
	FromPage *page
}

type updatedPage struct {
	Page      *page
	UpdatedAt string
	Updates   map[string]*update // update type -> update
}

// updatesTmplData stores the data that we pass to the updates.tmpl to render the page
type updatesTmplData struct {
	User         *user.User
	UpdatedPages []*updatedPage
}

// updatesPage serves the updates page.
var updatesPage = newPageWithOptions(
	"/updates/",
	updatesRenderer,
	append(baseTmpls,
		"tmpl/updatesPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"),
	newPageOptions{RequireLogin: true})

// updatesRenderer renders the updates page.
func updatesRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data updatesTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the updates
	data.UpdatedPages = make([]*updatedPage, 0)
	updatedPagesMap := make(map[int64]*updatedPage)
	pageMap := make(map[int64]*page)
	userMap := make(map[int64]*dbUser)
	var buffer bytes.Buffer
	query := fmt.Sprintf(`
		SELECT contextPageId,updatedAt,type,count,fromUserId,fromPageId
		FROM updates
		WHERE userId=%d AND seen=0
		ORDER BY updatedAt DESC
		LIMIT 50`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var u update
		var pageId, fromUserId, fromPageId int64
		var updateType, updatedAt string
		err := rows.Scan(
			&pageId,
			&updatedAt,
			&updateType,
			&u.Count,
			&fromUserId,
			&fromPageId)
		if err != nil {
			return fmt.Errorf("failed to scan an update: %v", err)
		}
		// Create/get the updated page.
		curPage, ok := updatedPagesMap[pageId]
		if !ok {
			curPage = &updatedPage{}
			curPage.Page, ok = pageMap[pageId]
			if !ok {
				curPage.Page = &page{PageId: pageId}
				pageMap[pageId] = curPage.Page
			}
			curPage.Updates = make(map[string]*update)
			curPage.UpdatedAt = updatedAt
			updatedPagesMap[pageId] = curPage
			data.UpdatedPages = append(data.UpdatedPages, curPage)
			buffer.WriteString(fmt.Sprintf("%d,", pageId))
		}

		// Create/get the current update.
		curUpdate, ok := curPage.Updates[updateType]
		if !ok {
			curUpdate = &u
			curPage.Updates[updateType] = curUpdate
		} else {
			curUpdate.Count += u.Count
		}

		// If there is a user, proces it.
		if fromUserId > 0 && curUpdate.FromUser == nil {
			curUser, ok := userMap[fromUserId]
			if !ok {
				curUser = &dbUser{Id: fromUserId}
				userMap[fromUserId] = curUser
			}
			curUpdate.FromUser = curUser
		}

		// If there is a fromPage, process it.
		if fromPageId > 0 && curUpdate.FromPage == nil {
			fromPage, ok := pageMap[fromPageId]
			if !ok {
				fromPage = &page{PageId: fromPageId}
				pageMap[fromPageId] = fromPage
				buffer.WriteString(fmt.Sprintf("%d,", fromPageId))
			}
			curUpdate.FromPage = fromPage
		}
		return nil
	})
	if err != nil {
		c.Errorf("error while loading updates: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load pages.
	pageIds := strings.TrimRight(buffer.String(), ",")
	if pageIds != "" {
		query = fmt.Sprintf(`
			SELECT pageId,privacyKey,title,alias
			FROM pages
			WHERE isCurrentEdit AND deletedBy=0 AND pageId IN (%s)`, pageIds)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var p page
			err := rows.Scan(
				&p.PageId,
				&p.PrivacyKey,
				&p.Title,
				&p.Alias)
			if err != nil {
				return fmt.Errorf("failed to scan a page: %v", err)
			}
			pageMap[p.PageId].Title = p.Title
			pageMap[p.PageId].Alias = p.Alias
			return nil
		})
		if err != nil {
			c.Errorf("error while loading pages: %v", err)
			return pages.InternalErrorWith(err)
		}
	}

	// Load the names for all users.
	err = loadUsersInfo(c, userMap)
	if err != nil {
		c.Errorf("error while loading user names: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load updates count.
	data.User.UpdateCount, err = loadUpdateCount(c, data.User.Id)
	if err != nil {
		c.Errorf("Couldn't retrieve updates count: %v", err)
	}

	c.Inc("updates_page_served_success")
	return pages.StatusOK(data)
}
