// userHelpers.go contains the dbUser struct as well as helpful functions.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// loadUpdateCount returns the number of unseen updates the given user has.
func loadUpdateCount(c sessions.Context, userId int64) (int, error) {
	var updateCount int
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(newCount), 0)
		FROM updates
		WHERE userId=%d`, userId)
	_, err := database.QueryRowSql(c, query, &updateCount)
	return updateCount, err
}

// loadUserGroups loads all the group names this user belongs to.
func loadUserGroups(c sessions.Context, u *user.User) error {
	u.GroupIds = make([]string, 0)
	query := fmt.Sprintf(`
		SELECT groupId
		FROM groupMembers
		WHERE userId=%d`, u.Id)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var groupId int64
		err := rows.Scan(&groupId)
		if err != nil {
			return fmt.Errorf("failed to scan for a member: %v", err)
		}
		u.GroupIds = append(u.GroupIds, fmt.Sprintf("%d", groupId))
		return nil
	})
	return err
}

// loadGroupNames loads the names and other info for the groups in the map
func loadGroupNames(c sessions.Context, u *user.User, groupMap map[int64]*core.Group) error {
	// Make sure all user groups are in the map
	for _, idStr := range u.GroupIds {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if _, ok := groupMap[id]; !ok {
			groupMap[id] = &core.Group{Id: id}
		}
	}

	// Create the group string
	groupCondition := "FALSE"
	if len(groupMap) > 0 {
		var buffer bytes.Buffer
		for id, _ := range groupMap {
			buffer.WriteString(fmt.Sprintf("%d,", id))
		}
		bufferStr := buffer.String()
		groupCondition = fmt.Sprintf("id IN (%s)", bufferStr[0:len(bufferStr)-1])
	}

	// Load names
	query := fmt.Sprintf(`
		SELECT id,name
		FROM groups
		WHERE %s OR isVisible`, groupCondition)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var g core.Group
		err := rows.Scan(&g.Id, &g.Name)
		if err != nil {
			return fmt.Errorf("failed to scan for a group: %v", err)
		}
		if _, ok := groupMap[g.Id]; !ok {
			groupMap[g.Id] = &g
		} else {
			groupMap[g.Id].Name = g.Name
		}
		return nil
	})
	return err
}

// loadDomains loads the domains for the given page and adds them to the map
func loadDomains(c sessions.Context, u *user.User, page *core.Page, pageMap map[int64]*core.Page, groupMap map[int64]*core.Group) error {
	query := fmt.Sprintf(`
		SELECT g.id,g.name,g.alias,g.createdAt,g.rootPageId
		FROM groups AS g
		JOIN pageDomainPairs as pd
		ON (g.id=pd.domainId)
		WHERE isDomain AND pd.pageId=%d`, page.PageId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var g core.Group
		err := rows.Scan(&g.Id, &g.Name, &g.Alias, &g.CreatedAt, &g.RootPageId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		groupMap[g.Id] = &g
		page.DomainIds = append(page.DomainIds, fmt.Sprintf("%d", g.Id))
		if _, ok := pageMap[g.RootPageId]; !ok {
			pageMap[g.RootPageId] = &core.Page{PageId: g.RootPageId}
		}
		return nil
	})
	return err
}
