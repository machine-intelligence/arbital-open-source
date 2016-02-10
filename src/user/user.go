// Package user manages information about the current user.
package user

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
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
	Id             int64  `json:"id,string"`
	FbUserId       int64  `json:"fbUserId,string"`
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
