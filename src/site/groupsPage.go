// groupsPage.go serves the groups template.
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

type member struct {
	dbUser

	CanAddMembers bool
	CanAdmin      bool
}

type group struct {
	// Populated from db.
	Name string

	// Optionally populated.
	Members    []*member
	UserMember *member // member obj corresponding to the active user
}

// groupsTmplData stores the data that we pass to the template to render the page
type groupsTmplData struct {
	User   *user.User
	Groups []*group
}

// groupsPage serves the recent pages page.
var groupsPage = newPage(
	"/groups/",
	groupsRenderer,
	append(baseTmpls,
		"tmpl/groupsPage.tmpl", "tmpl/pageHelpers.tmpl", "tmpl/navbar.tmpl", "tmpl/footer.tmpl"))

// groupsRenderer renders the page page.
func groupsRenderer(w http.ResponseWriter, r *http.Request, u *user.User) *pages.Result {
	var err error
	var data groupsTmplData
	data.User = u
	c := sessions.NewContext(r)

	// Load the groups and members
	groupMap := make(map[string]*group)
	data.Groups = make([]*group, 0, 50)
	query := fmt.Sprintf(`
		SELECT g.name,m.userId,u.firstName,u.lastName,m.canAddMembers,m.canAdmin
		FROM groups AS g
		LEFT JOIN (
			SELECT userId,groupName,canAddMembers,canAdmin
			FROM groupMembers
		) AS m
		ON (g.name=m.groupName)
		LEFT JOIN (
			SELECT id,firstName,lastName
			FROM users
		) AS u
		ON (m.userId=u.id)
		WHERE g.name IN (SELECT groupName FROM groupMembers WHERE userId=%d)`, data.User.Id)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var g group
		var m member
		err := rows.Scan(
			&g.Name,
			&m.Id,
			&m.FirstName,
			&m.LastName,
			&m.CanAddMembers,
			&m.CanAdmin)
		if err != nil {
			return fmt.Errorf("failed to scan a group member: %v", err)
		}

		// Add group
		curGroup := groupMap[g.Name]
		if curGroup == nil {
			curGroup = &g
			groupMap[g.Name] = curGroup
			data.Groups = append(data.Groups, curGroup)
		}
		// Add member
		curGroup.Members = append(curGroup.Members, &m)
		if m.Id == data.User.Id {
			curGroup.UserMember = &m
		}
		return nil
	})
	if err != nil {
		c.Errorf("error while loading group members: %v", err)
		return pages.InternalErrorWith(err)
	}

	funcMap := template.FuncMap{}
	c.Inc("groups_page_served_success")
	return pages.StatusOK(data).AddFuncMap(funcMap)
}
