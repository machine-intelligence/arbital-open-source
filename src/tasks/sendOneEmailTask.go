// sendOneEmailTask.go sends one email, with no delay before the task can run again
package tasks

import (
	"fmt"

	"appengine/mail"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// SendOneEmailTask is the object that's put into the daemon queue.
type SendOneEmailTask struct {
	UserId string
}

func (task SendOneEmailTask) Tag() string {
	return "sendOneEmail"
}

// Check if this task is valid, and we can safely execute it.
func (task SendOneEmailTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("User id has to be set: %v", task.UserId)
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task SendOneEmailTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Infof("==== SEND EMAIL START ====")
	defer c.Infof("==== SEND EMAIL COMPLETED ====")

	// Update database first, even though we might fail to send the email. This
	// way we definitely won't accidentally email a person twice.
	statement := db.NewStatement(`
		UPDATE users
		SET updateEmailSentAt=NOW()
		WHERE id=?`)
	_, err = statement.Exec(task.UserId)
	if err != nil {
		return 0, fmt.Errorf("failed to update updateEmailSentAt: %v", err)
	}

	emailData, err := core.LoadUpdateEmail(db, task.UserId)
	if err != nil {
		return 0, fmt.Errorf("Failed to load email text: %v", err)
	}

	// If emailData is nil, that means there is no update to send, so just return
	if emailData == nil {
		c.Infof("Nothing to send")
		return
	}

	if emailData.UpdateEmailAddress == "" || emailData.UpdateEmailText == "" {
		return 0, fmt.Errorf("Email is empty")
	}

	if sessions.Live {

		// Mark loaded updates as emailed
		updateIds := make([]interface{}, 0)
		for _, row := range emailData.UpdateRows {
			updateIds = append(updateIds, row.Id)
		}
		statement := database.NewQuery(`
			UPDATE updates
			SET emailed=true
			WHERE id IN`).AddArgsGroup(updateIds).ToStatement(db)
		_, err = statement.Exec()
		if err != nil {
			return 0, fmt.Errorf("Couldn't update updates as emailed: %v", err)
		}

		// Create mail message
		subject := fmt.Sprintf("%d new updates on Arbital", emailData.UpdateCount)
		msg := &mail.Message{
			Sender:   "alexei@arbital.com",
			To:       []string{emailData.UpdateEmailAddress},
			Bcc:      []string{"alexei@arbital.com"},
			Subject:  subject,
			HTMLBody: emailData.UpdateEmailText,
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

	c.Inc("email_send_success")
	c.Infof("Email sent!")

	return
}
