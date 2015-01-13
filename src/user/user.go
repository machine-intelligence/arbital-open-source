// Package user manages information about the current user.
package user

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/twitter"
)

var (
	userKey   = "user" // key for session storage
	EmptyUser = User{}
	fakeUser  = User{Twitter: twitter.TwitterUser{ScreenName: "Fakey fakeson", Id: 1, ProfileURL: "http://sophosnews.files.wordpress.com/2013/09/fake-sign_thumb.jpg"}}
)

// User holds information about a user of the app.
// Note: this structure is also stored in a cookie.
type User struct {
	Twitter      twitter.TwitterUser   // info loaded from corresponding twitter account
	TwitterCreds *sessions.Credentials // Twitter credentials
}

// Currency returns an English description of the user's currency.
func (u User) CurrencyString() string {
	return u.Currency.String()
}

// LoadUser returns user object from session, if available.
func LoadUser(r *http.Request) (*User, error) {
	c := sessions.NewContext(r)
	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}
	if s.Values[userKey] == nil {
		c.Debugf("no user in session\n")
		return nil, nil
	}
	user := s.Values[userKey].(*User)
	if *user == EmptyUser {
		c.Inc("empty_user_fail")
		return nil, fmt.Errorf("empty user in session")
	}
	return user, nil
}

// ParseUser returns a new user object from a io.ReadCloser.
//
// The io.ReadCloser might e.g. be a HTTP response body.
func ParseUser(rc io.ReadCloser) (*User, error) {
	var user User
	err := json.NewDecoder(rc).Decode(&user.Twitter)
	if err != nil {
		return nil, fmt.Errorf("Error decoding the user: %v", err)
	}
	return &user, nil
}

// Save stores the user in the session.
func (user *User) Save(w http.ResponseWriter, r *http.Request) error {
	s, err := sessions.GetSession(r)
	if err != nil {
		return fmt.Errorf("couldn't get session: %v", err)
	}

	s.Values[userKey] = user
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save user to session: %v", err)
	}
	return nil
}

// BecomeFakeUser sets the current user's cookie to a static fake profile.
func BecomeFakeUser(w http.ResponseWriter, r *http.Request) error {
	c := sessions.NewContext(r)
	if sessions.Live {
		m := "BecomeFakeUser was called on Live, which is a very bad idea\n"
		c.Criticalf(m)
		return fmt.Errorf(m)
	}
	err := sessions.FakeCreds.Save(w, r)
	if err != nil {
		return fmt.Errorf("failed to save fake creds: %v", err)
	}
	err = fakeUser.Save(w, r)
	if err != nil {
		return fmt.Errorf("failed to save fake user: %v", err)
	}
	return nil
}

func init() {
	gob.Register(&User{})
}
