// filterPage.go serves the filter template.
package site

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// filterTmplData stores the data that we pass to the template to render the page
type filterTmplData struct {
	commonPageData
	LimitCount int

	Author             *core.User
	IsSubscribedToUser bool
}

// filterPage serves the recent pages page.
var filterPage = newPageWithOptions(
	"/filter/",
	filterRenderer,
	append(baseTmpls,
		"tmpl/filterPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl",
		"tmpl/footer.tmpl", "tmpl/angular.tmpl.js"),
	newPageOptions{LoadUserGroups: true})

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
		data.Author = &core.User{}
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
	pageIds := make([]string, 0, 50)
	data.GroupMap = make(map[int64]*group)
	data.PageMap = make(map[int64]*core.Page)
	query := fmt.Sprintf(`
		SELECT p.pageId,p.edit,p.title,p.alias,p.privacyKey,p.groupId
		FROM pages AS p
		WHERE (p.privacyKey=0 OR p.creatorId=%d) AND isCurrentEdit AND p.deletedBy=0 AND
			(p.groupId=0 OR p.groupId IN (SELECT groupId FROM groupMembers WHERE userId=%[1]d)) %s
		ORDER BY p.createdAt DESC
		LIMIT %d`, data.User.Id, userConstraint, data.LimitCount)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var p core.Page
		err := rows.Scan(
			&p.PageId,
			&p.Edit,
			&p.Title,
			&p.Alias,
			&p.PrivacyKey,
			&p.GroupId)
		if err != nil {
			return fmt.Errorf("failed to scan a page: %v", err)
		}

		pageIds = append(pageIds, fmt.Sprintf("%d", p.PageId))
		data.PageMap[p.PageId] = &p
		return nil
	})
	if err != nil {
		c.Errorf("error while loading pages: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load tags.
	/*err = loadParents(c, data.PageMap, data.User.Id)
	if err != nil {
		c.Errorf("Couldn't retrieve page parents: %v", err)
		return pages.InternalErrorWith(err)
	}*/

	// Load auxillary data.
	err = loadAuxPageData(c, data.User.Id, data.PageMap, nil)
	if err != nil {
		c.Errorf("error while loading aux data: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load all the groups.
	err = loadGroupNames(c, u, data.GroupMap)
	if err != nil {
		c.Errorf("Couldn't load group names: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{}
	c.Inc("pages_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
