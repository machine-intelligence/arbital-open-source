// tick.go updates all the computed values in our database.
package tasks

import (
	"database/sql"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	tickPeriod = 5 * 60 // 5 minutes
)

// TickTask is the object that's put into the daemon queue.
type TickTask struct {
}

func (task TickTask) Tag() string {
	return "tick"
}

// Check if this task is valid, and we can safely execute it.
func (task TickTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task TickTask) Execute(db *database.DB) (delay int, err error) {
	delay = tickPeriod
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	query := database.NewQuery(`
		UPDATE pageInfos AS pi
		SET pi.viewCount=(
			SELECT COUNT(DISTINCT userId)
			FROM visits AS v
			WHERE v.pageId=pi.pageId
		)`).ToStatement(db)
	if _, err := query.Exec(); err != nil {
		c.Errorf("Failed to update view count: %v", err)
	}

	c.Infof("==== TICK START ====")
	defer c.Infof("==== TICK COMPLETED ====")
	return
}

func processQuestion(c sessions.Context, rows *sql.Rows) error {
	return nil
}
