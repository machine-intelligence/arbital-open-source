// sendFeedbackEmailTask.go sends a feedback email
package tasks

import (
	"fmt"

	"appengine/mail"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// SendFeedbackEmailTask is the object that's put into the daemon queue.
type SendFeedbackEmailTask struct {
	UserId    int64
	UserEmail string
	Text      string
}

// Check if this task is valid, and we can safely execute it.
func (task *SendFeedbackEmailTask) IsValid() error {
	if task.UserId <= 0 {
		return fmt.Errorf("User id has to be set")
	}
	if task.Text == "" {
		return fmt.Errorf("Text has to be set")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *SendFeedbackEmailTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== SEND FEEDBACK START ====")
	defer c.Debugf("==== SEND FEEDBACK COMPLETED ====")

	if sessions.Live {
		// Create mail message
		msg := &mail.Message{
			Sender:  "alexei@arbital.com",
			To:      []string{"alexei@arbital.com"},
			Cc:      []string{task.UserEmail},
			Subject: fmt.Sprintf("Arbital feedback (user #%d)", task.UserId),
			Body:    task.Text,
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

	c.Inc("feedback_send_success")
	c.Debugf("Feedback sent!")

	return
}
