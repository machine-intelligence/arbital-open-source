// sendTestEmailHandler.go displays the test email page
package site

import (
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
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
	var u *core.CurrentUser
	u, err = core.LoadCurrentUser(w, r, db)
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
	statement.Exec(u.ID)

	// Mark all updates as new and not emailed, for testing
	statement = db.NewStatement(`
		UPDATE updates
		SET seen=FALSE,emailed=0
		WHERE userId=?`)
	statement.Exec(u.ID)

	emailData, err := core.LoadUpdateEmail(db, u.ID)
	if err != nil {
		fmt.Fprintf(w, "Loading email failed: %v", err)
		return
	}

	if emailData == nil || emailData.UpdateEmailAddress == "" || emailData.UpdateEmailText == "" {
		fmt.Fprintf(w, "Email is empty")
		return
	}

	fmt.Fprintf(w, emailData.UpdateEmailText)
}
