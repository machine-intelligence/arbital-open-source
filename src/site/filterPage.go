// filterPage.go serves the filter template.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// filterTmplData stores the data that we pass to the template to render the page
type filterTmplData struct {
	User       *user.User
	Pages      []*page
	LimitCount int

	Author             *dbUser
	IsSubscribedToUser bool
}

// filterPage serves the recent pages page.
var filterPage = newPage(
	"/filter/",
	filterRenderer,
	append(baseTmpls,
		"tmpl/filterPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// filterRenderer renders the page page.
func filterRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data filterTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Check what parameters are passed in, since that'll change what pages we search for.
	q := r.URL.Query()
	// Check parameter limiting the number of pages returned
	data.LimitCount = 50
	recentParam := q.Get("recent")
	if recentParam != "" {
		data.LimitCount, _ = strconv.Atoi(recentParam)
		if data.LimitCount > 200 {
			data.LimitCount = 200
		}
	}
	// Check parameter limiting the user/creator of the pages
	var throwaway int
	userConstraint := ""
	userParam := q.Get("user")
	if userParam != "" {
		data.Author = &dbUser{}
		query := fmt.Sprintf(`
			SELECT id,firstName,lastName
			FROM users
			WHERE id=%s`, userParam)
		_, err = database.QueryRowSql(c, query,
			&data.Author.Id, &data.Author.FirstName, &data.Author.LastName)
		if err != nil {
			c.Errorf("Couldn't retrieve user: %v", err)
			return pages.BadRequestWith(err)
		}

		query = fmt.Sprintf(`
			SELECT 1
			FROM subscriptions
			WHERE userId=%d AND toUserId=%s`, data.User.Id, userParam)
		data.IsSubscribedToUser, err = database.QueryRowSql(c, query, &throwaway)
		if err != nil {
			c.Errorf("Couldn't retrieve subscription: %v", err)
			return pages.BadRequestWith(err)
		}

		userConstraint = fmt.Sprintf("AND creatorId=%s", userParam)
	}

	// Load the pages
	pageMap := make(map[int64]*page)
	pageIds := make([]string, 0, 50)
	data.Pages = make([]*page, 0, 50)
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.title,p.alias,p.privacyKey,p.groupName
		FROM pages AS p
		WHERE (p.privacyKey=0 OR p.creatorId=%d) AND isCurrentEdit AND p.deletedBy=0 AND
			(p.groupName="" OR p.groupName IN (SELECT groupName FROM groupMembers WHERE userId=%[1]d)) %s
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.User.Id, userConstraint, data.LimitCount)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p page
		err := rows.Scan(
			&p.PageId,
			&p.Edit,
			&p.Title,
			&p.Alias,
			&p.PrivacyKey,
			&p.Group.Name)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}

		pageMap[p.PageId] = &p
		pageIds = append(pageIds, fmt.Sprintf("%d", p.PageId))
		data.Pages = append(data.Pages, &p)
		return nil
	})
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load tags.
	/*err = loadParents(c, pageMap, data.User.Id)
	if err != nil {
		c.Errorf("Couldn't retrieve page parents: %v", err)
		return pages.InternalErrorWith(err)
	}*/

	// Load likes.
	err = loadLikes(c, data.User.Id, pageMap)
	if err != nil {
		c.Errorf("Couldn't retrieve page likes: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{}
	c.Inc("pages_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
