// userHelpers.go contains the dbUser struct as well as helpful functions.
package site

import (
	"fmt"
	"strconv"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

// loadUpdateCount returns the number of unseen updates the given user has.
func loadUpdateCount(db *database.DB, userId int64) (int, error) {
	var updateCount int
	row := db.NewStatement(`
		SELECT COALESCE(SUM(newCount), 0)
		FROM updates
		WHERE userId=?`).QueryRow(userId)
	_, err := row.Scan(&updateCount)
	return updateCount, err
}

// loadUserGroups loads all the group names this user belongs to.
func loadUserGroups(db *database.DB, u *user.User) error {
	u.GroupIds = make([]string, 0)
	rows := db.NewStatement(`
		SELECT groupId
		FROM groupMembers
		WHERE userId=?`).Query(u.Id)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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
func loadGroupNames(db *database.DB, u *user.User, groupMap map[int64]*core.Group) error {
	// Make sure all user groups are in the map
	for _, idStr := range u.GroupIds {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if _, ok := groupMap[id]; !ok {
			groupMap[id] = &core.Group{Id: id}
		}
	}

	// Create the group string
	groupCondition := "FALSE"
	groupIds := make([]interface{}, 0, len(groupMap))
	if len(groupMap) > 0 {
		for id, _ := range groupMap {
			groupIds = append(groupIds, id)
		}
		groupCondition = "id IN " + database.InArgsPlaceholder(len(groupIds))
	}

	// Load names
	rows := db.NewStatement(`
		SELECT id,name
		FROM groups
		WHERE ` + groupCondition + ` OR isVisible`).Query(groupIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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
func loadDomains(db *database.DB, u *user.User, page *core.Page, pageMap map[int64]*core.Page, groupMap map[int64]*core.Group) error {
	rows := db.NewStatement(`
		SELECT g.id,g.name,g.alias,g.createdAt,g.rootPageId
		FROM groups AS g
		JOIN pageDomainPairs AS pd
		ON (g.id=pd.domainId)
		WHERE isDomain AND pd.pageId=?`).Query(page.PageId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
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
