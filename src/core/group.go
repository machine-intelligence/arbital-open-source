// group.go contains all the stuff about groups
package core

import (
	"fmt"
	"strconv"

	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

type Member struct {
	UserId        int64 `json:"userId,string"`
	CanAddMembers bool  `json:"canAddMembers"`
	CanAdmin      bool  `json:"canAdmin"`
}

type Group struct {
	Id         int64  `json:"id,string"`
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	IsVisible  bool   `json:"isVisible"`
	RootPageId int64  `json:"rootPageId,string"`
	CreatedAt  string `json:"createdAt"`

	// Optionally populated.
	Members []*Member `json:"members"`
	// Member obj corresponding to the active user
	UserMember *Member `json:"userMember"`
}

// LoadUserGroups loads all the group names this user belongs to.
func LoadUserGroups(db *database.DB, u *user.User) error {
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

// LoadGroupNames loads the names and other info for the groups in the map
func LoadGroupNames(db *database.DB, u *user.User, groupMap map[int64]*Group) error {
	// Make sure all user groups are in the map
	for _, idStr := range u.GroupIds {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if _, ok := groupMap[id]; !ok {
			groupMap[id] = &Group{Id: id}
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
		SELECT id,name,alias
		FROM groups
		WHERE ` + groupCondition + ` OR isVisible`).Query(groupIds...)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var g Group
		err := rows.Scan(&g.Id, &g.Name, &g.Alias)
		if err != nil {
			return fmt.Errorf("failed to scan for a group: %v", err)
		}
		if _, ok := groupMap[g.Id]; !ok {
			groupMap[g.Id] = &g
		} else {
			// TODO: Nope, nope, nope! Can't do this again. Figure out if we ever have
			// a group that comes in here with preloaded info that we have to preserve.
			// If so, let's refactor it so that this function is the one that loads all
			// the data for groups. That way we can just replace the group, instead of
			// doing setting each variable.
			groupMap[g.Id].Name = g.Name
			groupMap[g.Id].Alias = g.Alias
		}
		return nil
	})
	return err
}

// LoadDomains loads the domains for the given page and adds them to the map
func LoadDomains(db *database.DB, u *user.User, page *Page, pageMap map[int64]*Page, groupMap map[int64]*Group) error {
	rows := db.NewStatement(`
		SELECT g.id,g.name,g.alias,g.createdAt,g.rootPageId
		FROM groups AS g
		JOIN pageDomainPairs AS pd
		ON (g.id=pd.domainId)
		WHERE isDomain AND pd.pageId=?`).Query(page.PageId)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var g Group
		err := rows.Scan(&g.Id, &g.Name, &g.Alias, &g.CreatedAt, &g.RootPageId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		groupMap[g.Id] = &g
		page.DomainIds = append(page.DomainIds, fmt.Sprintf("%d", g.Id))
		if _, ok := pageMap[g.RootPageId]; !ok {
			pageMap[g.RootPageId] = &Page{PageId: g.RootPageId}
		}
		return nil
	})
	return err
}
