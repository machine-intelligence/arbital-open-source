// atMentionUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// AtMentionUpdateTask is the object that's put into the daemon queue.
type AtMentionUpdateTask struct {
	// User who performed an action, e.g. creating a comment
	UserID string

	// User who was mentioned
	MentionedUserID string

	// Grouping key. One of these has to set. We'll group all updates by this key
	// to show in one panel.
	GroupByPageID string
	GroupByUserID string

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageID string
}

func (task AtMentionUpdateTask) Tag() string {
	return "atMentionUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task AtMentionUpdateTask) IsValid() error {
	if !core.IsIDValid(task.UserID) {
		return fmt.Errorf("UserId has to be set")
	} else if !core.IsIDValid(task.MentionedUserID) {
		return fmt.Errorf("MentionedUserId has to be set")
	}

	groupByCount := 0
	if core.IsIDValid(task.GroupByPageID) {
		groupByCount++
	}
	if core.IsIDValid(task.GroupByUserID) {
		groupByCount++
	}
	if groupByCount != 1 {
		return fmt.Errorf("Exactly one GroupBy... has to be set")
	}

	if !core.IsIDValid(task.GoToPageID) {
		return fmt.Errorf("GoToPageId has to be set")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task AtMentionUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		c.Errorf("Invalid @mention update task: %s", err)
		return -1, err
	}

	// Check if the user id is valid
	rows := database.NewQuery(`
		SELECT id
		FROM users`).Add(`
		WHERE id=?`, task.MentionedUserID).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to scan for subscriptions: %v", err)
		}
		if userID == task.UserID {
			return nil
		}

		// Insert new update
		hashmap := make(map[string]interface{})
		hashmap["userId"] = userID
		hashmap["byUserId"] = task.UserID
		hashmap["type"] = core.AtMentionUpdateType
		hashmap["groupByPageId"] = task.GroupByPageID
		hashmap["groupByUserId"] = task.GroupByUserID
		hashmap["goToPageId"] = task.GoToPageID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("updates", hashmap)
		if _, err = statement.Exec(); err != nil {
			c.Inc("new_update_fail")
			c.Errorf("Couldn't create new update: %v", err)
		}
		return nil
	})
	if err != nil {
		c.Errorf("Couldn't process users: %v", err)
		return -1, err
	}
	return 0, nil
}
