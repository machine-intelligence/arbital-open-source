// Package user manages information about the current user.
package user

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"time"

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
	sessionKey = "arbitalSession" // key for session storage

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
	FbUserId       string `json:"fbUserId"`
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
	UpdateCount int      `json:"updateCount"`
	GroupIds    []string `json:"groupIds"`
}

type CookieSession struct {
	Email string
	// Randomly generated string
	Random string
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

// Store user's email in a cookie
func SaveEmailCookie(w http.ResponseWriter, r *http.Request, email string) error {
	s, err := sessions.GetSession(r)
	if err != nil {
		return fmt.Errorf("Couldn't get session: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	s.Values[sessionKey] = CookieSession{email, fmt.Sprintf("%d", rand.Int63())}
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("Failed to save user to session: %v", err)
	}
	return nil
}

// loadUserFromDb tries to load the current user's info from the database. If
// there is no data in the DB, but the user is logged in through AppEngine,
// a new record is created.
func loadUserFromDb(r *http.Request, db *database.DB) (*User, error) {
	// Load email from the cookie
	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get session: %v", err)
	}

	var cookie *CookieSession
	if cookieStruct, ok := s.Values[sessionKey]; !ok {
		return nil, nil
	} else {
		cookie = cookieStruct.(*CookieSession)
	}

	var u User
	row := db.NewStatement(`
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,karma,emailFrequency,emailThreshold,inviteCode,ignoreMathjax
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	_, err = row.Scan(&u.Id, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin, &u.Karma,
		&u.EmailFrequency, &u.EmailThreshold, &u.InviteCode, &u.IgnoreMathjax)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	}
	u.MaxKarmaLock = GetMaxKarmaLock(u.Karma)
	return &u, nil
}

// LoadUser returns user object corresponding to logged in user. First, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database. If the user is not logged in, we return a partially filled
// User object.
// A user object is returned iff there is no error.
func LoadUser(r *http.Request, db *database.DB) (userPtr *User, err error) {
	userPtr, err = loadUserFromDb(r, db)
	if err != nil {
		return
	} else if userPtr == nil {
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
	// Need to register this so it can be stored inside a cookie
	gob.Register(&CookieSession{})
}

const (
	Base31Chars             = "0123456789bcdfghjklmnpqrstvwxyz"
	Base36Chars             = "0123456789abcdefghijklmnopqrstuvwxyz"
	Base31CharsForFirstChar = "0123456789"
)

// Replace a rune at a specific index in a string
func replaceAtIndex(db *database.DB, in string, r rune, i int) string {
	//	db.C.Debugf("in: %v", in)
	//	db.C.Debugf("r: %v", r)
	//	db.C.Debugf("i: %v", i)
	out := []rune(in)
	//	db.C.Debugf("out: %v", out)
	out[i] = r
	//	db.C.Debugf("out: %v", out)
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
// Since we decided that ids must begin with a digit, only allow characters 0-9 for the first character index
func GetNextBase31Char(db *database.DB, char rune, isFirstChar bool) (rune, bool, error) {
	validChars := Base31Chars
	if isFirstChar {
		validChars = Base31CharsForFirstChar
	}
	index := strings.Index(validChars, strings.ToLower(string(char)))
	//db.C.Debugf("index: %v", index)
	if index < 0 {
		return '0', false, fmt.Errorf("invalid character")
	}
	if index < len(validChars)-1 {
		nextChar := rune(validChars[index+1])
		//		db.C.Debugf("nextChar: %v", nextChar)
		return nextChar, false, nil
	} else {
		nextChar := rune(validChars[0])
		//		db.C.Debugf("nextChar: %v", nextChar)
		return nextChar, true, nil
	}
}

// Increment a base31 Id string
func IncrementBase31Id(db *database.DB, previousId string) (string, error) {
	// Add 1 to the base36 value, skipping vowels
	// Start at the last character in the Id string, carrying the 1 as many times as necessary
	nextAvailableId := previousId
	//	db.C.Debugf("nextAvailableId: %v", nextAvailableId)
	index := len(nextAvailableId) - 1
	//	db.C.Debugf("index: %v", index)
	//	db.C.Debugf("nextAvailableId[index]: %v", nextAvailableId[index])
	var newChar rune
	var err error
	processNextChar := true
	for processNextChar {
		// If we need to carry the 1 all the way to the beginning, then add a 1 at the beginning of the string
		if index < 0 {
			nextAvailableId = "1" + nextAvailableId
			processNextChar = false
		} else {
			// Increment the character at the current index in the Id string
			newChar, processNextChar, err = GetNextBase31Char(db, rune(nextAvailableId[index]), index == 0)
			//			db.C.Debugf("newChar: %v", newChar)
			//			db.C.Debugf("processNextChar: %v", processNextChar)
			if err != nil {
				return "", fmt.Errorf("Error processing id: %v", err)
			}
			nextAvailableId = replaceAtIndex(db, nextAvailableId, newChar, index)
			//			db.C.Debugf("nextAvailableId: %v", nextAvailableId)
			index = index - 1
			//			db.C.Debugf("index: %v", index)
		}
	}

	//	db.C.Debugf("nextAvailableId: %v", nextAvailableId)
	return nextAvailableId, nil
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
	/*
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
	*/

	row := db.NewStatement(`
SELECT MAX(pageId)
FROM (
SELECT pageId
FROM pageInfos
UNION 
SELECT id
FROM users
) AS combined
WHERE char_length(pageId) = 
(
SELECT MAX(char_length(pageId))
FROM (
SELECT pageId
FROM pageInfos
UNION 
SELECT id
FROM users
) AS combined2
    )
		`).QueryRow()
	_, err := row.Scan(&highestUsedId)
	if err != nil {
		return "", fmt.Errorf("Couldn't load id: %v", err)
	}

	nextAvailableId, err := IncrementBase31Id(db, highestUsedId)

	return nextAvailableId, err
}
