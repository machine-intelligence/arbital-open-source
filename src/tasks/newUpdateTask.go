// newUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// NewUpdateTask is the object that's put into the daemon queue.
type NewUpdateTask struct {
	// User who performed an action, e.g. creating a comment
	UserId     string
	UpdateType string

	// Grouping key. One of these has to set. We'll group all updates by this key
	// to show in one panel.
	GroupByPageId string
	GroupByUserId string

	// We'll notify the users who are subscribed to this page id (also could be a
	// user id, group id, domain id)
	SubscribedToId string

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageId string
}

// Check if this task is valid, and we can safely execute it.
func (task *NewUpdateTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("User id has to be set")
	} else if task.UpdateType == "" {
		return fmt.Errorf("Update type has to be set")
	} else if !core.IsIdValid(task.SubscribedToId) {
		return fmt.Errorf("SubscibedTo id has to be set")
	}

	groupByCount := 0
	if core.IsIdValid(task.GroupByPageId) {
		groupByCount++
	}
	if core.IsIdValid(task.GroupByUserId) {
		groupByCount++
	}
	if groupByCount != 1 {
		return fmt.Errorf("Exactly one GroupBy... has to be set")
	}

	if !core.IsIdValid(task.GoToPageId) {
		return fmt.Errorf("GoToPageId has to be set")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *NewUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		c.Errorf("Invalid new update task: %s", err)
		return -1, err
	}

	// Iterate through all users who are subscribed to this page/comment.
	rows := database.NewQuery(`
		SELECT userId
		FROM subscriptions
		WHERE toId=?`, task.SubscribedToId).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId string
		err := rows.Scan(&userId)
		if err != nil {
			return fmt.Errorf("failed to scan for subscriptions: %v", err)
		}
		if userId == task.UserId {
			return nil
		}

		// Check if we already have a similar update.
		var previousUpdateId string
		var exists bool
		newCountValue := 1
		row := db.NewStatement(`
			SELECT id
			FROM updates
			WHERE userId=? AND byUserId=? AND type=? AND newCount>0 AND
				groupByPageId=? AND groupByUserId=? AND
				subscribedToId=? AND goToPageId=?
			ORDER BY createdAt DESC
			LIMIT 1`).QueryRow(userId, task.UserId, task.UpdateType,
			task.GroupByPageId, task.GroupByUserId,
			task.SubscribedToId, task.GoToPageId)
		exists, err = row.Scan(&previousUpdateId)
		if err != nil {
			return fmt.Errorf("failed to check for existing update: %v", err)
		}
		if exists {
			// If we already have an update like this, don't count this one
			newCountValue = 0
		}

		// Insert new update
		hashmap := make(map[string]interface{})
		hashmap["userId"] = userId
		hashmap["byUserId"] = task.UserId
		hashmap["type"] = task.UpdateType
		hashmap["groupByPageId"] = task.GroupByPageId
		hashmap["groupByUserId"] = task.GroupByUserId
		hashmap["subscribedToId"] = task.SubscribedToId
		hashmap["goToPageId"] = task.GoToPageId
		hashmap["createdAt"] = database.Now()
		hashmap["newCount"] = newCountValue
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
