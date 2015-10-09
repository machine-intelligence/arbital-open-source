// atMentionUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/database"
)

// AtMentionTask is the object that's put into the daemon queue.
type AtMentionUpdateTask struct {
	// User who performed an action, e.g. creating a comment
	UserId     int64
	UpdateType string

	// User who was mentioned
	MentionedUserId     int64

	// Grouping key. One of these has to set. We'll group all updates by this key
	// to show in one panel.
	GroupByPageId int64
	GroupByUserId int64

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageId int64
}

// Check if this task is valid, and we can safely execute it.
func (task *AtMentionUpdateTask) IsValid() error {
	if task.UserId <= 0 {
		return fmt.Errorf("User id has to be set")
	} else if task.UpdateType == "" {
		return fmt.Errorf("Update type has to be set")
	}

	groupByCount := 0
	if task.GroupByPageId > 0 {
		groupByCount++
	}
	if task.GroupByUserId > 0 {
		groupByCount++
	}
	if groupByCount != 1 {
		return fmt.Errorf("Exactly one GroupBy... has to be set")
	}

	if task.GoToPageId <= 0 {
		return fmt.Errorf("GoToPageId has to be set")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *AtMentionUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		c.Errorf("Invalid @mention update task: %s", err)
		return -1, err
	}

	// Figure out the subscriptions query constraint.
	whereClause := database.NewQuery("")
	whereClause.Add("WHERE id=?", task.MentionedUserId)

	// Check if the user id is valid
	rows := database.NewQuery(`
		SELECT id
		FROM users`).AddPart(whereClause).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId int64
		err := rows.Scan(&userId)
		if err != nil {
			return fmt.Errorf("failed to scan for subscriptions: %v", err)
		}
		if userId == task.UserId {
			return nil
		}

		// Insert new update
		hashmap := make(map[string]interface{})
		hashmap["userId"] = userId
		hashmap["byUserId"] = task.UserId
		hashmap["type"] = task.UpdateType
		hashmap["groupByPageId"] = task.GroupByPageId
		hashmap["groupByUserId"] = task.GroupByUserId
		hashmap["goToPageId"] = task.GoToPageId
		hashmap["createdAt"] = database.Now()
		hashmap["newCount"] = 1
		statement := db.NewInsertStatement("updates", hashmap)
		if _, err = statement.Exec(); err != nil {
			c.Inc("new_update_fail")
			c.Errorf("Couldn't create new update: %v", err)
		}
		return nil
	})
	if err != nil {
		c.Errorf("Couldn't process subscriptions: %v", err)
		return -1, err
	}
	return 0, nil
}
