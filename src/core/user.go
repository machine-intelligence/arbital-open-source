// user.go contains all the stuff about user
package core

import (
	"database/sql"
	"fmt"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	// Karma requirements to perform various actions
	CommentKarmaReq            = -5
	LikeKarmaReq               = 0
	PrivatePageKarmaReq        = 5
	VoteKarmaReq               = 10
	AddParentKarmaReq          = 20
	CreateAliasKarmaReq        = 50
	EditPageKarmaReq           = 0 //50 // edit wiki page
	DeleteParentKarmaReq       = 100
	KarmaLockKarmaReq          = 100
	ChangeSortChildrenKarmaReq = 100
	ChangeAliasKarmaReq        = 200
	DeletePageKarmaReq         = 500
	DashlessAliasKarmaReq      = 1000
)

// User has information about a user from the users table.
type User struct {
	Id        int64  `json:"id,string"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// True if the current user is subscribed to this user
	IsSubscribed bool `json:"isSubscribed"`
}

func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// GetUserUrl returns URL for looking at recently created pages by the given user.
func GetUserUrl(userId int64) string {
	return fmt.Sprintf("/filter?user=%d", userId)
}

// LoadUsers loads user information (like name) for each user in the given map.
func LoadUsers(c sessions.Context, userMap map[int64]*User) error {
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
		var u User
		err := rows.Scan(&u.Id, &u.FirstName, &u.LastName)
		if err != nil {
			return fmt.Errorf("failed to scan for user: %v", err)
		}
		*userMap[u.Id] = u
		return nil
	})
	return err
}
