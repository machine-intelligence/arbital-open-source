// groupsPage.go serves the groups template.
package site

import (
	"database/sql"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// groupsTmplData stores the data that we pass to the template to render the page
type groupsTmplData struct {
	commonPageData
}

// groupsPage serves the recent pages page.
var groupsPage = newPage(
	"/groups/",
	groupsRenderer,
	append(baseTmpls,
		"tmpl/groupsPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/angular.tmpl.js",
		"tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// groupsRenderer renders the page page.
func groupsRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data groupsTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the groups and members
	data.UserMap = make(map[int64]*core.User)
	data.GroupMap = make(map[int64]*core.Group)
	query := fmt.Sprintf(`
		SELECT g.id,g.name,m.userId,m.canAddMembers,m.canAdmin
		FROM groups AS g
		LEFT JOIN (
			SELECT userId,groupId,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (g.id=m.groupId)
		WHERE g.id IN (SELECT groupId FROM groupMembers WHERE userId=%d)`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var g core.Group
		var m core.Member
		err := rows.Scan(
			&g.Id,
			&g.Name,
			&m.UserId,
			&m.CanAddMembers,
			&m.CanAdmin)
		if err != nil {
			return fmt.Errorf("failed to scan a group member: %v", err)
		}

		// Add group
		curGroup := data.GroupMap[g.Id]
		if curGroup == nil {
			curGroup = &g
			data.GroupMap[g.Id] = curGroup
		}

		// Add member
		curGroup.Members = append(curGroup.Members, &m)
		if m.UserId == data.User.Id {
			curGroup.UserMember = &m
		}
		data.UserMap[m.UserId] = &core.User{Id: m.UserId}
		return nil
	})
	if err != nil {
		c.Errorf("error while loading group members: %v", err)
		return pages.InternalErrorWith(err)
	}

	// Load all the users.
	err = core.LoadUsers(c, data.UserMap)
	if err != nil {
		c.Errorf("error while loading users: %v", err)
		return pages.InternalErrorWith(err)
	}

	c.Inc("groups_page_served_success")
	return pages.StatusOK(data)
}
