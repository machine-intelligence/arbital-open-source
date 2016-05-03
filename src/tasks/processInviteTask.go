// processInviteTask.go sends a invite email
package tasks

import (
	"fmt"

	"appengine/mail"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// ProcessInviteTask is the object that's put into the daemon queue.
type ProcessInviteTask struct {
	FromUserId string
	DomainId   string
	ToEmail    string
}

func (task ProcessInviteTask) Tag() string {
	return "processInvite"
}

// Check if this task is valid, and we can safely execute it.
func (task ProcessInviteTask) IsValid() error {
	if !core.IsIdValid(task.FromUserId) {
		return fmt.Errorf("Invalid FromUserId")
	}
	if !core.IsIdValid(task.DomainId) && task.DomainId != "" {
		return fmt.Errorf("Invalid DomainId")
	}
	if task.ToEmail == "" {
		return fmt.Errorf("Invalid ToEmail")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task ProcessInviteTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== SEND INVITE START ====")
	defer c.Debugf("==== SEND INVITE COMPLETED ====")

	wherePart := database.NewQuery(`WHERE fromUserId=?`, task.FromUserId).Add(`
		AND domainId=?`, task.DomainId).Add(`
		AND toEmail=?`, task.ToEmail).Add(`
		AND emailSentAt=0`)
	invites, err := core.LoadInvitesWhere(db, wherePart)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load invites: %v", err)
	} else if len(invites) <= 0 {
		return
	}

	// Load the necessary info about the invite
	var senderFirstName, senderLastName, domainTitle string
	row := database.NewQuery(`
			SELECT u.firstName,u.lastName,p.title
			FROM users AS u
			JOIN pages AS p
			ON (u.id=?`, task.FromUserId).Add(`
				AND p.pageId=?`, task.DomainId).Add(`
				AND p.isLiveEdit)`).ToStatement(db).QueryRow()
	exists, err := row.Scan(&senderFirstName, &senderLastName, &domainTitle)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load data: %v", err)
	} else if !exists {
		return
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Create new invite
		hashmap := make(map[string]interface{})
		hashmap["fromUserId"] = task.FromUserId
		hashmap["domainId"] = task.DomainId
		hashmap["toEmail"] = task.ToEmail
		hashmap["emailSentAt"] = database.Now()
		statement := db.NewInsertStatement("invites", hashmap, "emailSentAt").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't add row to invites table", err
		}

		senderName := senderFirstName + " " + senderLastName
		bodyText := fmt.Sprintf(`
			Hello! %s just sent you an invite to Arbital,
			an ambitious effort to solve online discussion.\n\n
			This invite gives you the permission to create and edit pages in the %s domain.\n\n
			Visit https://arbital.com/signup to create your account.\n\n
			We're excited to have you with us!\n\n
			—Team Arbital`, senderName, domainTitle)

		if sessions.Live {
			// Create mail message
			msg := &mail.Message{
				Sender:  "alexei@arbital.com",
				To:      []string{task.ToEmail},
				Bcc:     []string{"alexei@arbital.com"},
				Subject: fmt.Sprintf("Arbital invite from %s [Domain: %s])", senderName, domainTitle),
				Body:    bodyText,
			}

			// Ship it!
			err = mail.Send(c, msg)
			if err != nil {
				c.Inc("email_send_fail")
				return "Couldn't send email", err
			}
		} else {
			// If not live, then do nothing, for now
		}
		return "", nil
	})
	if errMessage != "" {
		return -1, fmt.Errorf(errMessage, err)
	}

	c.Inc("invite_send_success")
	c.Debugf("Invite sent!")

	return
}
