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
	userKey  = "user" // key for session storage
	fakeUser = User{Id: -1, Email: "fake@fake.com", FirstName: "Dr.", LastName: "Fake"}
)

// User holds information about a user of the app.
// Note: this structure is also stored in a cookie.
type User struct {
	Id         int64
	Email      string
	FirstName  string
	LastName   string
	IsLoggedIn bool
	IsAdmin    bool
	LoginLink  string
	LogoutLink string
}

// Get the currently logged in user from App Engine. The users database is
// updated if this user is newly created.
func current(r *http.Request, c sessions.Context) (*User, error) {
	appEngineUser := user.Current(c)
	if appEngineUser == nil {
		return nil, nil
	}

	var u User
	sql := fmt.Sprintf("SELECT id,email,firstName,lastName FROM users WHERE email='%s'", appEngineUser.Email)
	exists, err := database.QueryRowSql(c, sql, &u.Id, &u.Email, &u.FirstName, &u.LastName)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		// Add new user
		c.Debugf("User not found. Creating a new one: %s", appEngineUser.Email)
		dbUser := make(database.InsertMap)
		dbUser["email"] = appEngineUser.Email
		dbUser["firstName"] = ""
		dbUser["lastName"] = ""
		dbUser["isAdmin"] = appEngineUser.Admin
		dbUser["createdAt"] = database.Now()
		dbUser["lastWebsiteVisit"] = database.Now()
		sql := database.GetInsertSql("users", dbUser)
		err = database.ExecuteSql(c, sql)
		if err != nil {
			return nil, fmt.Errorf("Couldn't creat a new user: %v", err)
		}
		u.Email = appEngineUser.Email
		u.IsAdmin = appEngineUser.Admin
	}
	u.IsLoggedIn = true
	return &u, nil
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
	return nil
}

// LoadUser returns user object from session, if available. Otherwise, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database and update the cookie. If the user is not logged in, we
// return a partially filled User object.
// A user object is returned iff there is no error.
func LoadUser(w http.ResponseWriter, r *http.Request) (userPtr *User, err error) {
	c := sessions.NewContext(r)
	s, err2 := sessions.GetSession(r)
	if err2 != nil {
		err = fmt.Errorf("failed to get session: %v", err2)
		return
	}

	if s.Values[userKey] != nil {
		c.Debugf("loading user from session")
		userPtr = s.Values[userKey].(*User)
	} else {
		c.Debugf("no user in session, checking app engine")
		userPtr, err = current(r, c)
		if err != nil {
			return
		} else if userPtr != nil {
			userPtr.Save(w, r)
		} else {
			userPtr = &User{}
		}
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
