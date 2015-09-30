// tick.go updates all the computed values in our database.
package tasks

import (
	"database/sql"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	tickPeriod = 3600
)

// TickTask is the object that's put into the daemon queue.
type TickTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *TickTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *TickTask) Execute(db *database.DB) (delay int, err error) {
	delay = tickPeriod
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== TICK START ====")
	defer c.Debugf("==== TICK COMPLETED ====")
	return
}

func processQuestion(c sessions.Context, rows *sql.Rows) error {
	return nil
}
