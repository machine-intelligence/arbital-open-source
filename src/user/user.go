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
	sessionKey = "arbitalSession2" // key for session storage

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

	// If the user isn't logged in, this is set to their unique session id
	SessionId string `json:"-"`

	// Computed variables
	UpdateCount int      `json:"updateCount"`
	GroupIds    []string `json:"groupIds"`
}

type CookieSession struct {
	Email     string
	SessionId string

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
	s.Values[sessionKey] = CookieSession{Email: email, Random: fmt.Sprintf("%d", rand.Int63())}
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("Failed to save user to session: %v", err)
	}
	return nil
}

// Store a unique id in a cookie so we can track session
func saveSessionCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	s, err := sessions.GetSession(r)
	if err != nil {
		return "", fmt.Errorf("Couldn't get session: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	sessionId := fmt.Sprintf("%d", rand.Int63())
	s.Values[sessionKey] = CookieSession{
		SessionId: sessionId,
		Random:    fmt.Sprintf("%d", rand.Int63()),
	}
	err = s.Save(r, w)
	if err != nil {
		return "", fmt.Errorf("Failed to save user to session: %v", err)
	}
	return sessionId, nil
}

// loadUserFromDb tries to load the current user's info from the database. If
// there is no data in the DB, but the user is logged in through AppEngine,
// a new record is created.
func loadUserFromDb(w http.ResponseWriter, r *http.Request, db *database.DB) (*User, error) {
	// Load email from the cookie
	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get session: %v", err)
	}

	var cookie *CookieSession
	if cookieStruct, ok := s.Values[sessionKey]; !ok {
		sessionId, err := saveSessionCookie(w, r)
		return &User{SessionId: sessionId}, err
	} else {
		cookie = cookieStruct.(*CookieSession)
	}
	if cookie.Email == "" {
		return &User{SessionId: cookie.SessionId}, err
	}

	var u User
	row := db.NewStatement(`
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,karma,emailFrequency,emailThreshold,inviteCode,ignoreMathjax
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	exists, err := row.Scan(&u.Id, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin, &u.Karma,
		&u.EmailFrequency, &u.EmailThreshold, &u.InviteCode, &u.IgnoreMathjax)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find that email in DB")
	}
	u.MaxKarmaLock = GetMaxKarmaLock(u.Karma)
	return &u, nil
}

// LoadUser returns user object corresponding to logged in user. First, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database. If the user is not logged in, we return a partially filled
// User object.
// A user object is returned iff there is no error.
func LoadUser(w http.ResponseWriter, r *http.Request, db *database.DB) (userPtr *User, err error) {
	userPtr, err = loadUserFromDb(w, r, db)
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
	Base31CharsForFirstChar = "0123456789"
)

// Replace a rune at a specific index in a string
func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

// Get the next highest base36 character, without vowels
// Returns the character, and true if it wrapped around to 0
// Since we decided that ids must begin with a digit, only allow characters 0-9 for the first character index
func GetNextBase31Char(c sessions.Context, char rune, isFirstChar bool) (rune, bool, error) {
	validChars := Base31Chars
	if isFirstChar {
		validChars = Base31CharsForFirstChar
	}
	index := strings.Index(validChars, strings.ToLower(string(char)))
	if index < 0 {
		return '0', false, fmt.Errorf("invalid character")
	}
	if index < len(validChars)-1 {
		nextChar := rune(validChars[index+1])
		return nextChar, false, nil
	} else {
		nextChar := rune(validChars[0])
		return nextChar, true, nil
	}
}

// Increment a base31 Id string
func IncrementBase31Id(c sessions.Context, previousId string) (string, error) {
	// Add 1 to the base36 value, skipping vowels
	// Start at the last character in the Id string, carrying the 1 as many times as necessary
	nextAvailableId := previousId
	index := len(nextAvailableId) - 1
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
			newChar, processNextChar, err = GetNextBase31Char(c, rune(nextAvailableId[index]), index == 0)
			if err != nil {
				return "", fmt.Errorf("Error processing id: %v", err)
			}
			nextAvailableId = replaceAtIndex(nextAvailableId, newChar, index)
			index = index - 1
		}
	}

	return nextAvailableId, nil
}

// Call GetNextAvailableId in a new transaction
func GetNextAvailableIdInNewTransaction(db *database.DB) (string, error) {
	return db.Transaction(func(tx *database.Tx) (string, error) {
		return GetNextAvailableId(tx)
	})
}

// Get the next available base36 Id string that doesn't contain vowels
func GetNextAvailableId(tx *database.Tx) (string, error) {
	// Query for the highest used pageId or userId
	var highestUsedId string
	row := tx.NewTxStatement(`
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

	nextAvailableId, err := IncrementBase31Id(tx.DB.C, highestUsedId)

	return nextAvailableId, err
}
