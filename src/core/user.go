// user.go contains some user things.
package core

import (
	"fmt"

	"zanaduu3/src/database"
)

// User has a selection of the information about a user.
type User struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	LastWebsiteVisit string `json:"lastWebsiteVisit"`

	// Computed variables
	// True if the currently logged in user is subscribed to this user
	IsSubscribed bool `json:"isSubscribed"`
}

// AddUserIdToMap adds a new user with the given user id to the map if it's not
// in the map already.
// Returns the new/existing user.
func AddUserIDToMap(userID string, userMap map[string]*User) *User {
	if !IsIDValid(userID) {
		return nil
	}
	if u, ok := userMap[userID]; ok {
		return u
	}
	u := &User{ID: userID}
	userMap[userID] = u
	return u
}

// FullName returns user's first and last name.
func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// LoadUsers loads User information (like name) for each user in the given map.
func LoadUsers(db *database.DB, userMap map[string]*User, currentUserID string) error {
	if len(userMap) <= 0 {
		return nil
	}
	userIDs := make([]interface{}, 0, len(userMap))
	for id := range userMap {
		userIDs = append(userIDs, id)
	}

	rows := database.NewQuery(`
		SELECT u.id,u.firstName,u.lastName,u.lastWebsiteVisit,!ISNULL(s.userId)
		FROM (
			SELECT *
			FROM users
			WHERE id IN `).AddArgsGroup(userIDs).Add(`
		) AS u
		LEFT JOIN (
			SELECT *
			FROM subscriptions WHERE userId=?`, currentUserID).Add(`
		) AS s
		ON (u.id=s.toId)`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var u User
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.LastWebsiteVisit, &u.IsSubscribed)
		if err != nil {
			return fmt.Errorf("failed to scan for user: %v", err)
		}
		*userMap[u.ID] = u
		return nil
	})
	return err
}
func LoadUser(db *database.DB, userID string, currentUserID string) (*User, error) {
	user := &User{ID: userID}
	userMap := map[string]*User{userID: user}
	err := LoadUsers(db, userMap, currentUserID)
	return user, err
}

// LoadUserTrust returns the trust that the user has in all domains.
func LoadUserTrust(db *database.DB, userID string) (map[string]*Trust, error) {
	trustMap := make(map[string]*Trust)
	rows := database.NewQuery(`
		SELECT domainId,max(generalTrust),max(editTrust)
		FROM (
			SELECT ut.domainId AS domainId,ut.generalTrust AS generalTrust,ut.editTrust AS editTrust
			FROM userTrust AS ut
			WHERE ut.userId=?`, userID).Add(`
			UNION ALL
			SELECT d.pageId AS domainId,0 AS generalTrust,0 AS editTrust
			FROM domains AS d
		) AS u
		GROUP BY 1`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var trust Trust
		var domainID string
		err := rows.Scan(&domainID, &trust.GeneralTrust, &trust.EditTrust)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		trustMap[domainID] = &trust
		if trust.EditTrust >= ArbiterKarmaLevel {
			trust.Level = ArbiterTrustLevel
		} else if trust.EditTrust >= ReviewerKarmaLevel {
			trust.Level = ReviewerTrustLevel
		} else if trust.EditTrust >= BasicKarmaLevel {
			trust.Level = BasicTrustLevel
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error while loading userTrust: %v", err)
	}

	return trustMap, nil
}
