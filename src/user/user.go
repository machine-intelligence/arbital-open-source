// Package user manages information about the current user.
package user

import (
	"database/sql"
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
	// DB variables
	Id        int64
	Email     string
	FirstName string
	LastName  string
	IsAdmin   bool
	Karma     int

	// Computed variables
	IsLoggedIn  bool
	CurrentUrl  string
	LoginLink   string
	LogoutLink  string
	UpdateCount int
}

func (user *User) FullName() string {
	return user.FirstName + " " + user.LastName
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

// BecomeUserWithId allows an admin to pretend they are a specific user.
func (u *User) BecomeUserWithId(id string, c sessions.Context) error {
	query := fmt.Sprintf("SELECT id,email,firstName,lastName,isAdmin FROM users WHERE id=%s", id)
	exists, err := database.QueryRowSql(c, query, &u.Id, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin)
	if err != nil {
		return fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return fmt.Errorf("Couldn't find a user with id: %s", id)
	}
	return nil
}

// loadUserFromDb tries to load the current user's info from the database. If
// there is no data in the DB, but the user is logged in through AppEngine,
// a new record is created.
func loadUserFromDb(r *http.Request) (*User, error) {
	c := sessions.NewContext(r)
	appEngineUser := user.Current(c)
	if appEngineUser == nil {
		return nil, nil
	}

	var u User
	query := fmt.Sprintf("SELECT id,email,firstName,lastName,isAdmin,karma FROM users WHERE email='%s'", appEngineUser.Email)
	exists, err := database.QueryRowSql(c, query, &u.Id, &u.Email, &u.FirstName, &u.LastName, &u.IsAdmin, &u.Karma)
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

		var result sql.Result
		query := database.GetInsertSql("users", dbUser)
		result, err = database.ExecuteSql(c, query)
		if err != nil {
			return nil, fmt.Errorf("Couldn't create a new user: %v", err)
		}
		u.Id, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("Couldn't get last insert id for new user: %v", err)
		}
		u.Email = appEngineUser.Email
	}
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
func LoadUser(w http.ResponseWriter, r *http.Request) (userPtr *User, err error) {
	c := sessions.NewContext(r)
	userPtr, err = loadUserFromDb(r)
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
