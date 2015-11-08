// emailUpdatesTask.go sends an email to every user with the new updates
package tasks

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

const (
	// Run EmailUpdatesTask once every 5 minutes
	emailUpdatesPeriod = 60 * 5
)

// EmailUpdatesTask is the object that's put into the daemon queue.
type EmailUpdatesTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *EmailUpdatesTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *EmailUpdatesTask) Execute(db *database.DB) (delay int, err error) {
	delay = emailUpdatesPeriod
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== EMAIL UPDATES START ====")
	defer c.Debugf("==== EMAIL UPDATES COMPLETED SUCCESSFULLY ====")

	// Find all users who need emailing.
	rows := db.NewStatement(`
		SELECT id
		FROM users
		WHERE (DATEDIFF(NOW(),updateEmailSentAt)>=7 AND emailFrequency=?)
			OR (DATEDIFF(NOW(),updateEmailSentAt)>=1 AND emailFrequency=?)
			OR (DATEDIFF(NOW(),updateEmailSentAt)>=0 AND emailFrequency=?)`).Query(user.WeeklyEmailFrequency, user.DailyEmailFrequency, user.ImmediatelyEmailFrequency)

	err = rows.Process(emailUpdatesProcessUser)
	if err != nil {
		c.Errorf("ERROR: %v", err)
	}
	return
}

func emailUpdatesProcessUser(db *database.DB, rows *database.Rows) error {
	c := db.C

	var userId int64
	err := rows.Scan(&userId)
	if err != nil {
		return fmt.Errorf("failed to scan a user id: %v", err)
	}

	var task SendOneEmailTask
	task.UserId = userId
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	} else if err := Enqueue(c, task, "sendOneEmail"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return nil
}
