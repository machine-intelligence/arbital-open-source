// sendInviteTask.go sends a invite email
package tasks

import (
	"fmt"

	"google.golang.org/appengine/mail"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// SendInviteTask is the object that's put into the daemon queue.
type SendInviteTask struct {
	FromUserID string
	DomainID   string
	ToEmail    string
}

func (task SendInviteTask) Tag() string {
	return "sendInvite"
}

// Check if this task is valid, and we can safely execute it.
func (task SendInviteTask) IsValid() error {
	if !core.IsIDValid(task.FromUserID) {
		return fmt.Errorf("Invalid FromUserId")
	}
	if !core.IsIntIDValid(task.DomainID) {
		return fmt.Errorf("Invalid domain id: %v", task.DomainID)
	}
	if task.ToEmail == "" {
		return fmt.Errorf("Invalid ToEmail")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task SendInviteTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Infof("==== SEND INVITE START ====")
	defer c.Infof("==== SEND INVITE COMPLETED ====")

	senderUser, err := core.LoadUser(db, task.FromUserID, task.FromUserID)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load sender user info: %v", err)
	}

	domain, err := core.LoadDomainByID(db, task.DomainID)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load domain info: %v", err)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		bodyText := fmt.Sprintf(`
Hello!

%s just sent you an invite to Arbital, an ambitious effort to solve online discussion.

Visit https://arbital.com/signup to create your account.

We're excited to have you with us!

—Team Arbital`, senderUser.FullName(), domain.Alias)

		if sessions.Live {
			// Create mail message
			msg := &mail.Message{
				Sender:  "alexei@arbital.com",
				To:      []string{task.ToEmail},
				Bcc:     []string{"alexei@arbital.com"},
				Subject: fmt.Sprintf("Arbital invite from %s", senderUser.FullName()),
				Body:    bodyText,
			}

			// Ship it!
			err = mail.Send(c, msg)
			if err != nil {
				c.Inc("email_send_fail")
				return sessions.NewError("Couldn't send email", err)
			}
		} else {
			// If not live, then do nothing, for now
		}
		return nil
	})
	if err2 != nil {
		return -1, sessions.ToError(err2)
	}

	c.Inc("invite_send_success")
	c.Infof("Invite sent!")

	return
}
