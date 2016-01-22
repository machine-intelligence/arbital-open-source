// user.go contains all the stuff about user
package core

import (
	"fmt"

	"zanaduu3/src/database"
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

	// Enter the correct invite code to get karma
	CorrectInviteCode  = "TRUTH"
	CorrectInviteKarma = 200
)

// User has information about a user from the users table.
type User struct {
	Id               int64  `json:"id,string"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	LastWebsiteVisit string `json:"lastWebsiteVisit"`
	// True if the currently logged in user is subscribed to this user
	IsSubscribed bool `json:"isSubscribed"`
}

// FullName returns user's first and last name.
func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// GetUserUrl returns URL for looking at recently created pages by the given user.
func GetUserUrl(userId int64) string {
	return fmt.Sprintf("/user/%d", userId)
}

// IdsListFromUserMap returns a list of all user ids in the map.
func IdsListFromUserMap(userMap map[int64]*User) []interface{} {
	list := make([]interface{}, 0, len(userMap))
	for id, _ := range userMap {
		list = append(list, id)
	}
	return list
}

// LoadUsers loads user information (like name) for each user in the given map.
func LoadUsers(db *database.DB, userMap map[int64]*User, userId int64) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIds := make([]interface{}, 0, len(userMap))
	for id, _ := range userMap {
		userIds = append(userIds, id)
	}

	rows := database.NewQuery(`
		SELECT u.id,u.firstName,u.lastName,u.lastWebsiteVisit,!ISNULL(s.userId)
		FROM (
			SELECT *
			FROM users
			WHERE id IN `).AddArgsGroup(userIds).Add(`
		) AS u
		LEFT JOIN (
			SELECT *
			FROM subscriptions WHERE userId=?`, userId).Add(`
		) AS s
		ON (u.id=s.toId)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var u User
		err := rows.Scan(&u.Id, &u.FirstName, &u.LastName, &u.LastWebsiteVisit, &u.IsSubscribed)
		if err != nil {
			return fmt.Errorf("failed to scan for user: %v", err)
		}
		*userMap[u.Id] = u
		return nil
	})
	return err
}

// LoadUpdateCount returns the number of unseen updates the given user has.
func LoadUpdateCount(db *database.DB, userId int64) (int, error) {
	var updateCount int
	row := db.NewStatement(`
		SELECT COALESCE(SUM(newCount), 0)
		FROM updates
		WHERE userId=?`).QueryRow(userId)
	_, err := row.Scan(&updateCount)
	return updateCount, err
}
