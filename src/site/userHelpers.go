// userHelpers.go contains the dbUser struct as well as helpful functions.
package site

import (
	"database/sql"
	"fmt"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

const (
	// Karma requirements to perform various actions
	commentKarmaReq            = -5
	likeKarmaReq               = 0
	privatePageKarmaReq        = 5
	voteKarmaReq               = 10
	addParentKarmaReq          = 20
	createAliasKarmaReq        = 50
	editPageKarmaReq           = 0 //50 // edit wiki page
	deleteParentKarmaReq       = 100
	karmaLockKarmaReq          = 100
	changeSortChildrenKarmaReq = 100
	changeAliasKarmaReq        = 200
	deletePageKarmaReq         = 500
	dashlessAliasKarmaReq      = 1000
)

// dbUser has information about a user from the users table.
// We can't call this struct "user" since that collides with src/user.
type dbUser struct {
	// DB values.
	Id        int64  `json:"id,string"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// loadUsersInfo loads user information (like name) for each user in the given map.
func loadUsersInfo(c sessions.Context, userMap map[int64]*dbUser) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := make([]string, 0, len(userMap))
	for id, _ := range userMap {
		userIds = append(userIds, fmt.Sprintf("%d", id))
	}
	userIdsStr := strings.Join(userIds, ",")
	query := fmt.Sprintf(`
		SELECT id,firstName,lastName
		FROM users
		WHERE id IN (%s)`, userIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var u dbUser
		err := rows.Scan(&u.Id, &u.FirstName, &u.LastName)
		if err != nil {
			return fmt.Errorf("failed to scan for user: %v", err)
		}
		*userMap[u.Id] = u
		return nil
	})
	return err
}

// loadUpdateCount returns the number of unseen updates the given user has.
func loadUpdateCount(c sessions.Context, userId int64) (int, error) {
	var updateCount int
	query := fmt.Sprintf(`
		SELECT COALESCE(COUNT(distinct contextPageId), 0)
		FROM updates
		WHERE userId=%d AND seen=0`, userId)
	_, err := database.QueryRowSql(c, query, &updateCount)
	return updateCount, err
}

// loadUserGroups loads all the group names this user belongs to.
func loadUserGroups(c sessions.Context, u *user.User) error {
	// Load my groups.
	u.GroupNames = make([]string, 0)
	query := fmt.Sprintf(`
		SELECT groupName
		FROM groupMembers
		WHERE userId=%d`, u.Id)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var groupName string
		err := rows.Scan(&groupName)
		if err != nil {
			return fmt.Errorf("failed to scan for a member: %v", err)
		}
		u.GroupNames = append(u.GroupNames, groupName)
		return nil
	})
	return err
}

// getUserUrl returns URL for looking at recently created pages by the given user.
func getUserUrl(userId int64) string {
	return fmt.Sprintf("/filter?user=%d", userId)
}
