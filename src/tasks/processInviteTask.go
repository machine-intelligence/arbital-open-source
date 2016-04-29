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
	UserId  string
	EmailTo string
	Code    string
}

func (task ProcessInviteTask) Tag() string {
	return "processInvite"
}

// Check if this task is valid, and we can safely execute it.
func (task ProcessInviteTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("User id is not valid: '%v'", task.UserId)
	}
	if task.Code == "" {
		return fmt.Errorf("Code has to be set")
	}
	if task.EmailTo == "" {
		return fmt.Errorf("EmailTo has to be set")
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

	var senderFirstName, senderLastName, domainTitle string
	row := database.NewQuery(`
				SELECT u.firstName,u.lastName,p.title
				FROM users AS u
				JOIN invites AS i
				ON (i.code=?)`, task.Code).Add(`
				JOIN pages AS p
				ON (i.domainId=p.pageId AND p.isCurrentEdit)
				WHERE type=? AND alias=?`).ToStatement(db).QueryRow()
	exists, err := row.Scan(&senderFirstName, &senderLastName, &domainTitle)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load data: %v", err)
	} else if !exists {
		return
	}

	senderName := senderFirstName + " " + senderLastName
	bodyText := fmt.Sprintf(`
			Hello! %s just sent you an invite to Arbital,
			an ambitious effort to solve online discussion.\n\n
			Your code, %s, gives you permission edit in the %s domain.\n\n
			Visit www.arbital.com and sign up using your code, or
			if you already have an account, visit https://arbital.com/settings
			and enter your code there.\n\n
			We're excited to have you with us!\n\n
			—The Arbital Team`, senderName, task.Code, domainTitle)

	if sessions.Live {
		// Create mail message
		msg := &mail.Message{
			Sender:  "alexei@arbital.com",
			To:      []string{task.EmailTo},
			Bcc:     []string{"alexei@arbital.com"},
			Subject: fmt.Sprintf("Arbital invite from %s [Domain: %s])", senderName, domainTitle),
			Body:    bodyText,
		}

		// Ship it!
		err = mail.Send(c, msg)
		if err != nil {
			c.Inc("email_send_fail")
			return 0, fmt.Errorf("Couldn't send email: %v", err)
		}
	} else {
		// If not live, then do nothing, for now
	}

	c.Inc("invite_send_success")
	c.Debugf("Invite sent!")

	return
}
