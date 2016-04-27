// sendInviteEmailTask.go sends a invite email
package tasks

import (
	"fmt"

	"appengine/mail"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// SendInviteEmailTask is the object that's put into the daemon queue.
type SendInviteEmailTask struct {
	UserId        string
	InviteeEmails []string
	Code          string
}

func (task SendInviteEmailTask) Tag() string {
	return "sendInviteEmail"
}

// Check if this task is valid, and we can safely execute it.
func (task SendInviteEmailTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("User id has to be set: %v", task.UserId)
	}
	if task.Code == "" {
		return fmt.Errorf("Code can't be blank")
	}
	if len(task.InviteeEmails) == 0 {
		return fmt.Errorf("Address list can't be empty")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task SendInviteEmailTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== SEND INVITE START ====")
	defer c.Debugf("==== SEND INVITE COMPLETED ====")

	sender := "TODO: Load sender's name"

	bodyText := fmt.Sprintf(`
			Hello! %s (%s) just sent you an invite to Arbital,
			an ambitious effort to solve online discussion.\n\n
			Your code, %s, gives you permission edit in the %s domain.\n\n
			Visit www.arbital.com and sign up using your code, or
			if you already have an account, visit https://arbital.com/settings
			and enter your code there.\n\n
			We're excited to have you with us!\n\n
			—The Arbital Team`, sender, "TODO: Load sender's email", task.Code, "TODO: GET DOMAIN NAME")

	if sessions.Live {
		// Create mail message
		msg := &mail.Message{
			Sender:  "alexei@arbital.com",
			To:      task.InviteeEmails,
			Bcc:     []string{"alexei@arbital.com"},
			Subject: fmt.Sprintf("Arbital invite from %s [Domain: %s])", sender, "TODO: Get domain name"),
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
