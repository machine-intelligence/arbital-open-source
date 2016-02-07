// Package user manages information about the current user.
package user

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"appengine/user"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	DailyEmailFrequency       = "daily"
	WeeklyEmailFrequency      = "weekly"
	NeverEmailFrequency       = "never"
	ImmediatelyEmailFrequency = "immediately"
)

var (
	userKey = "user" // key for session storage

	// Highest karma lock a user can create is equal to their karma * this constant.
	MaxKarmaLockFraction = 0.8

	DefaultEmailFrequency = DailyEmailFrequency
	DefaultEmailThreshold = 3
)

// User holds information about a user of the app.
// Note: this structure is also stored in a cookie.
type User struct {
	// DB variables
	Id             string `json:"id"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	IsAdmin        bool   `json:"isAdmin"`
	Karma          int    `json:"karma"`
	MaxKarmaLock   int    `json:"maxKarmaLock"`
	EmailFrequency string `json:"emailFrequency"`
	EmailThreshold int    `json:"emailThreshold"`
	InviteCode     string `json:"inviteCode"`
	IgnoreMathjax  bool   `json:"ignoreMathjax"`

	// Computed variables
	IsLoggedIn  bool     `json:"isLoggedIn"`
	UpdateCount int      `json:"updateCount"`
	GroupIds    []string `json:"groupIds"`
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
func (user *User) IsMemberOfGroup(groupId string) bool {
	isMember := false
	oldGroupIdStr := groupId
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
		SELECT id,email,firstName,lastName,isAdmin,karma,emailFrequency,emailThreshold,inviteCode,ignoreMathjax
		FROM users
		WHERE email=?`).QueryRow(appEngineUser.Email)
	exists, err := row.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin, &u.Karma,
		&u.EmailFrequency, &u.EmailThreshold, &u.InviteCode, &u.IgnoreMathjax)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		// Add new user
		db.C.Debugf("User not found. Creating a new one: %s", appEngineUser.Email)

		u.Id, err = GetNextAvailableId(db)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get last insert id for new user: %v", err)
		}

		insertMap := make(database.InsertMap)
		insertMap["id"] = u.Id
		insertMap["email"] = appEngineUser.Email
		insertMap["firstName"] = ""
		insertMap["lastName"] = ""
		insertMap["isAdmin"] = appEngineUser.Admin
		insertMap["createdAt"] = database.Now()
		insertMap["lastWebsiteVisit"] = database.Now()
		insertMap["updateEmailSentAt"] = database.Now()
		insertMap["emailFrequency"] = DefaultEmailFrequency
		insertMap["emailThreshold"] = DefaultEmailThreshold
		insertMap["inviteCode"] = ""
		insertMap["ignoreMathjax"] = 0

		statement := db.NewInsertStatement("users", insertMap)
		_, err := statement.Exec()
		if err != nil {
			return nil, fmt.Errorf("Couldn't create a new user: %v", err)
		}
		u.Email = appEngineUser.Email
	}
	u.MaxKarmaLock = GetMaxKarmaLock(u.Karma)
	u.IsLoggedIn = u.FirstName != ""
	return &u, err
}

// GetLoginLink returns the link to log the user in
func GetLoginLink(c sessions.Context, continueUrl string) (string, error) {
	return user.LoginURL(c, continueUrl)
}

// GetLogoutLink returns the link to log the user out
func GetLogoutLink(c sessions.Context, continueUrl string) (string, error) {
	return user.LogoutURL(c, continueUrl)
}

// LoadUser returns user object corresponding to logged in user. First, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database. If the user is not logged in, we return a partially filled
// User object.
// A user object is returned iff there is no error.
func LoadUser(w http.ResponseWriter, r *http.Request, db *database.DB) (userPtr *User, err error) {
	userPtr, err = loadUserFromDb(db)
	if err != nil {
		return
	} else if userPtr != nil {
		userPtr.Save(w, r)
	} else {
		userPtr = &User{}
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

const (
	Base31Chars = "0123456789bcdfghjklmnpqrstvwxyz"
	Base36Chars = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// Replace a rune at a specific index in a string
func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

// Return true if the character is a vowel
func CharIsVowel(char rune) bool {
	switch []rune(strings.ToLower(string(char)))[0] {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// Get the next highest base36 character, without vowels
// Returns the character, and true if it wrapped around to 0
func GetNextBase31Char(char rune) (rune, bool, error) {
	index := strings.Index(Base31Chars, strings.ToLower(string(char)))
	if index < 0 {
		return '0', false, fmt.Errorf("invalid character")
	}
	if index < len(Base31Chars)-1 {
		nextChar := rune(Base31Chars[index+1])
		return nextChar, false, nil
	} else {
		nextChar := rune(Base31Chars[0])
		return nextChar, true, nil
	}
}

// Get the next available base36 Id string that doesn't contain vowels
func GetNextAvailableId(db *database.DB) (string, error) {
	// Query for the highest used pageId or userId
	var highestUsedId string
	/*
		row := db.NewStatement(`
			SELECT max(pageId)
			FROM pages
			WHERE 1
			`).QueryRow()
	*/
	row := db.NewStatement(`
SELECT MAX( pageId ) 
FROM (
SELECT pageId
FROM pages
UNION 
SELECT id
FROM users
) AS combined
		`).QueryRow()
	_, err := row.Scan(&highestUsedId)
	if err != nil {
		return "", fmt.Errorf("Couldn't load id: %v", err)
	}

	// Add 1 to the base36 value, skipping vowels
	// Start at the last character in the Id string, carrying the 1 as many times as necessary
	nextAvailableId := highestUsedId
	index := len(nextAvailableId)
	var newChar rune
	processNextChar := true
	for processNextChar {
		// Increment the character at the current index in the Id string
		newChar, processNextChar, err = GetNextBase31Char(rune(nextAvailableId[index]))
		if err != nil {
			return "", fmt.Errorf("Error processing id: %v", err)
		}
		replaceAtIndex(nextAvailableId, newChar, index)
		index = index - 1
		// If we need to carry the 1 all the way to the beginning, then add a 1 at the beginning of the string
		if index < 0 {
			nextAvailableId = "1" + nextAvailableId
			processNextChar = false
		}
	}

	return nextAvailableId, nil
}
