// sendTestEmailHandler.go displays the test email page
package site

import (
	"fmt"
	"net/http"

	"math/rand"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"

	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

func sendTestEmailHandler(w http.ResponseWriter, r *http.Request) {

	rand.Seed(time.Now().UnixNano())

	c := sessions.NewContext(r)
	fail := func(responseCode int, message string, err error) {
		c.Inc(fmt.Sprintf("%s-fail", r.URL.Path))
		c.Errorf("handlerWrapper: %s: %v", message, err)
		w.WriteHeader(responseCode)
		fmt.Fprintf(w, "%s", message)
	}

	// Recover from panic.
	defer func() {
		if sessions.Live {
			if r := recover(); r != nil {
				c.Errorf("%v", r)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", "Super serious error has occured. Super. Serious. Error.")
			}
		}
	}()
	// Open DB connection
	db, err := database.GetDB(c)
	if err != nil {
		fail(http.StatusInternalServerError, "Couldn't open DB", err)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r, db)
	if err != nil {
		fail(http.StatusInternalServerError, "Couldn't load user", err)
		return
	}

	if !u.IsAdmin {
		fail(http.StatusInternalServerError, "Have to be an admin", err)
		return
	}

	c.Debugf("In sendTestEmailHandler, starting task")

	// Set last email sent date to user created date, for testing
	statement := db.NewStatement(`
		UPDATE users
		SET updateEmailSentAt=createdAt
		WHERE id=?`)
	statement.Exec(u.Id)

	statement = db.NewStatement(`
		UPDATE updates
		SET newCount=1
		WHERE userId=?`)
	statement.Exec(u.Id)

	var resultData string

	_, resultData = tasks.SendTheEmail(db, u.Id, u.Email, u.EmailThreshold)

	fmt.Fprintf(w, resultData)
}
