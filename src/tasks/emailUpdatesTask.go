// emailUpdatesTask.go sends an email to every user with the new updates
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	// Run EmailUpdatesTask once every 5 minutes
	emailUpdatesPeriod = 60 * 5
)

// EmailUpdatesTask is the object that's put into the daemon queue.
type EmailUpdatesTask struct {
}

func (task EmailUpdatesTask) Tag() string {
	return "emailUpdates"
}

// Check if this task is valid, and we can safely execute it.
func (task EmailUpdatesTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task EmailUpdatesTask) Execute(db *database.DB) (delay int, err error) {
	delay = emailUpdatesPeriod
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Infof("==== EMAIL UPDATES START ====")
	defer c.Infof("==== EMAIL UPDATES COMPLETED ====")

	// For all the users that don't want emails, set their updates to 'emailed'
	statement := db.NewStatement(`
		UPDATE updates
		SET emailed=true
		WHERE userId IN (
			SELECT id
			FROM users
			WHERE emailFrequency=?
		)`)
	_, err = statement.Exec(core.NeverEmailFrequency)
	if err != nil {
		err = fmt.Errorf("Failed to update updates: %v", err)
		return
	}

	// NOTE: this query is tightly coupled with a query in update.go::LoadUpdateRows
	// Find all users who need emailing.
	rows := database.NewQuery(`
		SELECT u.id
		FROM users AS u
		JOIN (
			SELECT updates.userId,updates.createdAt,updates.type,updates.seen,updates.dismissed,updates.emailed,
				COALESCE((
					SELECT !isDeleted
					FROM pageInfos AS pi
					WHERE updates.goToPageId = pi.pageId
						AND`).AddPart(core.PageInfosFilterWithOptions(nil, &core.PageInfosOptions{Deleted: true})).Add(`
				), false) AS isGoToPageAlive
			FROM updates
		) AS up
		ON (u.id = up.userId)
		WHERE NOT up.seen
			AND NOT up.emailed
			AND NOT up.dismissed
			AND (DATEDIFF(up.createdAt,NOW()) >= 1 OR u.emailFrequency=?)`, core.ImmediatelyEmailFrequency).Add(`
			AND (up.isGoToPageAlive OR up.type IN`).AddArgsGroupStr(core.GetOkayToShowWhenGoToPageIsDeletedUpdateTypes()).Add(`)
		GROUP BY 1`).ToStatement(db).Query()
	err = rows.Process(emailUpdatesProcessUser)
	if err != nil {
		c.Errorf("Error while counting updates per user: %v", err)
	}
	return
}

func emailUpdatesProcessUser(db *database.DB, rows *database.Rows) error {
	c := db.C

	var userID string
	err := rows.Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to scan a user id: %v", err)
	}

	var task SendOneEmailTask
	task.UserID = userID
	if err := Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return nil
}
