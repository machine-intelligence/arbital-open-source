// sendTestEmailHandler.go displays the test email page
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

func sendTestEmailHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Open DB connection
	db, err := database.GetDB(c)
	if err != nil {
		fmt.Fprintf(w, "Failed to load database")
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(r, db)
	if err != nil {
		fmt.Fprintf(w, "Failed to load user")
		return
	}

	if !u.IsAdmin {
		fmt.Fprintf(w, "Have to be an admin")
		return
	}

	// Set last email sent date to user created date, for testing
	statement := db.NewStatement(`
		UPDATE users
		SET updateEmailSentAt=createdAt
		WHERE id=?`)
	statement.Exec(u.Id)

	// Mark all updates as new and not emailed, for testing
	statement = db.NewStatement(`
		UPDATE updates
		SET newCount=1,emailed=0
		WHERE userId=?`)
	statement.Exec(u.Id)

	emailData, err := core.LoadUpdateEmail(db, u.Id)
	if err != nil {
		fmt.Fprintf(w, "Loading email failed")
		return
	}

	if emailData.UpdateEmailAddress == "" || emailData.UpdateEmailText == "" {
		fmt.Fprintf(w, "Email is empty")
		return
	}

	fmt.Fprintf(w, emailData.UpdateEmailText)
}
