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
	Id               string `json:"id"`
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
func GetUserUrl(userId string) string {
	return fmt.Sprintf("/user/%s", userId)
}

// IdsListFromUserMap returns a list of all user ids in the map.
func IdsListFromUserMap(userMap map[string]*User) []interface{} {
	list := make([]interface{}, 0, len(userMap))
	for id, _ := range userMap {
		list = append(list, id)
	}
	return list
}

// LoadUsers loads user information (like name) for each user in the given map.
func LoadUsers(db *database.DB, userMap map[string]*User, userId string) error {
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
func LoadUpdateCount(db *database.DB, userId string) (int, error) {
	editTypes := []string{"pageEdit", "commentEdit"}

	var editUpdateCount int
	row := database.NewQuery(`
		SELECT COUNT(DISTINCT type, subscribedToId, byUserId)
		FROM updates
		WHERE unseen AND userId=?`, userId).Add(` AND type IN
	`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err := row.Scan(&editUpdateCount)
	if err != nil {
		return -1, err
	}

	var nonEditUpdateCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM updates
		WHERE unseen AND userId=?`, userId).Add(` AND type NOT IN
	`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err = row.Scan(&nonEditUpdateCount)
	if err != nil {
		return -1, err
	}

	return editUpdateCount + nonEditUpdateCount, err
}
