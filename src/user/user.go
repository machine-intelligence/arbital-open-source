// Package user manages information about the current user.
package user

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"appengine/user"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

var (
	userKey = "user" // key for session storage

	// Highest karma lock a user can create is equal to their karma * this constant.
	MaxKarmaLockFraction = 0.8

	DefaultEmailFrequency = "Daily"
	DefaultEmailThreshold = 3
)

// User holds information about a user of the app.
// Note: this structure is also stored in a cookie.
type User struct {
	// DB variables
	Id             int64  `json:"id,string"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	IsAdmin        bool   `json:"isAdmin"`
	Karma          int    `json:"karma"`
	MaxKarmaLock   int    `json:"maxKarmaLock"`
	EmailFrequency string `json:"emailFrequency"`
	EmailThreshold int    `json:"emailThreshold"`

	// Computed variables
	IsLoggedIn  bool     `json:"isLoggedIn"`
	CurrentUrl  string   `json:"currentUrl"`
	LoginLink   string   `json:"loginLink"`
	LogoutLink  string   `json:"logoutLink"`
	UpdateCount int      `json:"updateCount"`
	GroupIds    []string `json:"groupIds"`
	DomainAlias string   `json:"domainAlias"`
}

func (user *User) FullName() string {
	return user.FirstName + " " + user.LastName
}

// GetMaxKarmaLock returns the highest possible karma lock a user with the
// given amount of karma can create.
func GetMaxKarmaLock(karma int) int {
	return int(float64(karma) * MaxKarmaLockFraction)
}

// IsMemberOfGroup returns true iff the user is member of the given group.
// NOTE: we are assuming GroupIds have been loaded.
func (user *User) IsMemberOfGroup(groupId int64) bool {
	isMember := false
	oldGroupIdStr := fmt.Sprintf("%d", groupId)
	for _, groupIdStr := range user.GroupIds {
		if groupIdStr == oldGroupIdStr {
			isMember = true
			break
		}
	}
	return isMember
}

// Save stores the user in the session.
func (u *User) Save(w http.ResponseWriter, r *http.Request) error {
	/*s, err := sessions.GetSession(r)
	if err != nil {
		return fmt.Errorf("couldn't get session: %v", err)
	}

	s.Values[userKey] = u
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save user to session: %v", err)
	}*/
	return nil
}

// loadUserFromDb tries to load the current user's info from the database. If
// there is no data in the DB, but the user is logged in through AppEngine,
// a new record is created.
func loadUserFromDb(db *database.DB) (*User, error) {
	appEngineUser := user.Current(db.C)
	if appEngineUser == nil {
		return nil, nil
	}

	var u User
	row := db.NewStatement(`
		SELECT id,email,firstName,lastName,isAdmin,karma,emailFrequency,emailThreshold
		FROM users
		WHERE email=?`).QueryRow(appEngineUser.Email)
	exists, err := row.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin, &u.Karma,
		&u.EmailFrequency, &u.EmailThreshold)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		// Add new user
		db.C.Debugf("User not found. Creating a new one: %s", appEngineUser.Email)
		insertMap := make(database.InsertMap)
		insertMap["email"] = appEngineUser.Email
		insertMap["firstName"] = ""
		insertMap["lastName"] = ""
		insertMap["isAdmin"] = appEngineUser.Admin
		insertMap["createdAt"] = database.Now()
		insertMap["lastWebsiteVisit"] = database.Now()
		insertMap["updateEmailSentAt"] = database.Now()
		insertMap["emailFrequency"] = DefaultEmailFrequency
		insertMap["emailThreshold"] = DefaultEmailThreshold

		statement := db.NewInsertStatement("users", insertMap)
		result, err := statement.Exec()
		if err != nil {
			return nil, fmt.Errorf("Couldn't create a new user: %v", err)
		}
		u.Id, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("Couldn't get last insert id for new user: %v", err)
		}
		u.Email = appEngineUser.Email
	}
	u.MaxKarmaLock = GetMaxKarmaLock(u.Karma)
	u.IsLoggedIn = u.FirstName != ""
	return &u, err
}

// Set Login/Logout links for the given user object.
func setLinksForUser(r *http.Request, c sessions.Context, u *User) (err error) {
	u.LoginLink, err = user.LoginURL(c, r.URL.String())
	if err != nil {
		return fmt.Errorf("error getting login url: %v", err)
	}
	u.LogoutLink, err = user.LogoutURL(c, r.URL.String())
	if err != nil {
		return fmt.Errorf("error getting logout url: %v", err)
	}
	u.CurrentUrl = r.URL.String()
	return nil
}

// LoadUser returns user object corresponding to logged in user. First, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database. If the user is not logged in, we return a partially filled
// User object.
// A user object is returned iff there is no error.
func LoadUser(w http.ResponseWriter, r *http.Request, db *database.DB) (userPtr *User, err error) {
	c := sessions.NewContext(r)
	userPtr, err = loadUserFromDb(db)
	if err != nil {
		return
	} else if userPtr != nil {
		userPtr.Save(w, r)
	} else {
		userPtr = &User{}
	}
	if err = setLinksForUser(r, c, userPtr); err != nil {
		userPtr = nil
	}
	return
}

// ParseUser returns a new user object from a io.ReadCloser.
//
// The io.ReadCloser might e.g. be a HTTP response body.
func ParseUser(rc io.ReadCloser) (*User, error) {
	var user User
	err := json.NewDecoder(rc).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("Error decoding the user: %v", err)
	}
	return &user, nil
}

func init() {
	gob.Register(&User{})
}
