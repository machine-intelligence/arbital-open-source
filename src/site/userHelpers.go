// userHelpers.go contains the dbUser struct as well as helpful functions.
package site

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
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
func loadGroupNames(c sessions.Context, u *user.User, groupMap map[int64]*group) error {
	// Make sure all user groups are in the map
	for _, idStr := range u.GroupIds {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if _, ok := groupMap[id]; !ok {
			groupMap[id] = &group{Id: id}
		}
	}

	// Create the group string
	var buffer bytes.Buffer
	for id, _ := range groupMap {
		buffer.WriteString(fmt.Sprintf("%d,", id))
	}
	groupIdsStr := buffer.String()
	if len(groupIdsStr) >= 1 {
		groupIdsStr = groupIdsStr[0 : len(groupIdsStr)-1]
	}

	// Load names
	query := fmt.Sprintf(`
		SELECT id,name
		FROM groups
		WHERE id IN (%s)`, groupIdsStr)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var id int64
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			return fmt.Errorf("failed to scan for a group: %v", err)
		}
		groupMap[id].Name = name
		return nil
	})
	return err
}

// getUserUrl returns URL for looking at recently created pages by the given user.
func getUserUrl(userId int64) string {
	return fmt.Sprintf("/filter?user=%d", userId)
}
