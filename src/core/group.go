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

// LoadUserGroupIds loads all the group names this user belongs to.
func LoadUserGroupIds(db *database.DB, u *user.User) error {
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

// AddUserGroupIdsToPageMap adds user's groups to the page map so we can load them.
func AddUserGroupIdsToPageMap(u *user.User, pageMap map[int64]*Page) {
	for _, pageIdStr := range u.GroupIds {
		pageId, _ := strconv.ParseInt(pageIdStr, 10, 64)
		if _, ok := pageMap[pageId]; !ok {
			pageMap[pageId] = &Page{PageId: pageId}
		}
	}
}

// LoadDomainIds loads the domain ids for the given page and adds them to the map
func LoadDomainIds(db *database.DB, u *user.User, page *Page, pageMap map[int64]*Page) error {
	pageIdConstraint := database.NewQuery("")
	if page != nil {
		pageIdConstraint = database.NewQuery("AND pd.pageId=?", page.PageId)
	}
	rows := database.NewQuery(`
		SELECT p.pageId
		FROM pages AS p
		JOIN pageDomainPairs AS pd
		ON (p.pageId=pd.domainId)
		WHERE p.type="domain"`).AddPart(pageIdConstraint).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId int64
		err := rows.Scan(&domainId)
		if err != nil {
			return fmt.Errorf("failed to scan for a domain: %v", err)
		}
		page.DomainIds = append(page.DomainIds, fmt.Sprintf("%d", domainId))
		if _, ok := pageMap[domainId]; !ok {
			pageMap[domainId] = &Page{PageId: domainId}
		}
		return nil
	})
	return err
}
