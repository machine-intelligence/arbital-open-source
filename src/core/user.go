// user.go contains some user things.
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

// User has a selection of the information about a user.
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

// LoadUsers loads User information (like name) for each user in the given map.
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
