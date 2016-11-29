// user.go contains some user things.
package core

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// corePageData has data we load directly from the users and other tables.
type coreUserData struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	LastWebsiteVisit string `json:"lastWebsiteVisit"`

	// Computed variables
	// True if the currently logged in user is subscribed to this user
	IsSubscribed bool `json:"isSubscribed"`
}

// User has a selection of the information about a user.
type User struct {
	coreUserData

	// Which domains this user belongs to; map key is "domain id"
	DomainMembershipMap map[string]*DomainMember `json:"domainMembershipMap"`
}

// Return a new user object
func NewUser() *User {
	var u User
	u.DomainMembershipMap = make(map[string]*DomainMember)
	return &u
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
	u := NewUser()
	u.ID = userID
	userMap[userID] = u
	return u
}

// Returns domain id corresponding to this user.
func (u *User) MyDomainID() string {
	for _, dm := range u.DomainMembershipMap {
		if dm.DomainPageID == u.ID {
			return dm.DomainID
		}
	}
	return ""
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
		var u coreUserData
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.LastWebsiteVisit, &u.IsSubscribed)
		if err != nil {
			return fmt.Errorf("failed to scan for user: %v", err)
		}
		userMap[u.ID].coreUserData = u
		return nil
	})
	return err
}
func LoadUser(db *database.DB, userID string, currentUserID string) (*User, error) {
	userMap := make(map[string]*User)
	user := AddUserIDToMap(userID, userMap)
	err := LoadUsers(db, userMap, currentUserID)
	return user, err
}

// GetDomainMembershipRole returns the role the user has in the given domain.
// NOTE: we are assuming DomainMembershipMap has been loaded.
func (u *User) GetDomainMembershipRole(domainID string) string {
	if domainMember, ok := u.DomainMembershipMap[domainID]; ok {
		return domainMember.Role
	}
	return NoDomainRole
}

// LoadUserDomainMembership loads all the group names this user belongs to.
func LoadUserDomainMembership(db *database.DB, u *User, domainMap map[string]*Domain) error {
	u.DomainMembershipMap = make(map[string]*DomainMember)
	rows := db.NewStatement(`
		SELECT dm.domainId,dm.userId,dm.createdAt,dm.role,d.pageId
		FROM domainMembers AS dm
		JOIN domains AS d
		ON (dm.domainId=d.id)
		WHERE dm.userId=?`).Query(u.ID)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var dm DomainMember
		err := rows.Scan(&dm.DomainID, &dm.UserID, &dm.CreatedAt, &dm.Role, &dm.DomainPageID)
		if err != nil {
			return fmt.Errorf("failed to scan for a member: %v", err)
		}
		dm.CanCreateApprovedComments = RoleAtLeast(dm.Role, DefaultDomainRole)
		dm.CanApproveComments = RoleAtLeast(dm.Role, ReviewerDomainRole)
		u.DomainMembershipMap[dm.DomainID] = &dm
		if domainMap != nil {
			domainMap[dm.DomainID] = &Domain{ID: dm.DomainID}
		}
		return nil
	})
	return err
}

// NewUserDomainPage create a new page for a user and assigns it to a new domain.
func NewUserDomainPage(tx *database.Tx, u *CurrentUser, fullName, alias string) sessions.Error {
	url := GetEditPageFullURL("", u.ID)
	// NOTE: the tabbing/spacing is really important here since it gets preserved.
	// If we have 4 spaces, in Markdown that will start a code block.
	text := fmt.Sprintf(`
Automatically generated page for "%s" user.
If you are the owner, click [here to edit](%s).`, fullName, url)

	// Create a new domain
	var domainID string
	hashmap := make(database.InsertMap)
	hashmap["pageId"] = u.ID
	hashmap["alias"] = alias
	hashmap["createdBy"] = u.ID
	hashmap["createdAt"] = database.Now()
	hashmap["canUsersComment"] = true
	hashmap["canUsersProposeEdits"] = true
	statement := tx.DB.NewInsertStatement("domains", hashmap).WithTx(tx)
	if result, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't create a new domain row", err)
	} else {
		domainIDInt, err := result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get id of the new domain", err)
		}
		domainID = fmt.Sprintf("%d", domainIDInt)
	}

	// Add user to the domain
	hashmap = make(database.InsertMap)
	hashmap["userId"] = u.ID
	hashmap["domainId"] = domainID
	hashmap["createdAt"] = database.Now()
	hashmap["role"] = string(ReviewerDomainRole)
	statement = tx.DB.NewInsertStatement("domainMembers", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return sessions.NewError("Couldn't add user to the group", err)
	}

	// Create the new user page
	_, err := CreateNewPage(tx.DB, u, &CreateNewPageOptions{
		PageID:       u.ID,
		Alias:        alias,
		Type:         GroupPageType,
		Title:        fullName,
		Clickbait:    "Automatically generated page for " + fullName,
		Text:         text,
		EditDomainID: domainID,
		IsPublished:  true,
		Tx:           tx,
	})
	if err != nil {
		return sessions.NewError("Couldn't create a new page", err)
	}

	return nil
}
