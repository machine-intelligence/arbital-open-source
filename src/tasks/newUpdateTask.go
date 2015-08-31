// newUpdate.go inserts corresponding update.
package tasks

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// NewUpdateTask is the object that's put into the daemon queue.
type NewUpdateTask struct {
	// User who performed an action, e.g. creating a comment
	UserId     int64
	UpdateType string

	// Grouping key. One of these has to set. We'll group all updates by this key
	// to show in one panel.
	GroupByPageId int64
	GroupByUserId int64

	// Subscription check. One of these has to be set. We'll notify the users who
	// are subscribed to "this thing", e.g. this page id.
	SubscribedToPageId int64
	SubscribedToUserId int64

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageId int64
}

// Check if this task is valid, and we can safely execute it.
func (task *NewUpdateTask) IsValid() error {
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

	toCount := 0
	if task.SubscribedToPageId > 0 {
		toCount++
	}
	if task.SubscribedToUserId > 0 {
		toCount++
	}
	if toCount != 1 {
		return fmt.Errorf("Exactly one of 'SubscribedTo...' variables has to be set")
	}

	if task.GoToPageId <= 0 {
		return fmt.Errorf("GoToPageId has to be set")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *NewUpdateTask) Execute(c sessions.Context) (delay int, err error) {
	if err = task.IsValid(); err != nil {
		c.Errorf("Invalid new update task: %s", err)
		return -1, err
	}

	// Figure out the subscriptions query constraint.
	var whereClause string
	if task.SubscribedToPageId > 0 {
		whereClause = fmt.Sprintf("WHERE toPageId=%d", task.SubscribedToPageId)
	} else if task.SubscribedToUserId > 0 {
		whereClause = fmt.Sprintf("WHERE toUserId=%d", task.SubscribedToUserId)
	} else {
		return -1, err
	}

	// Iterate through all users who are subscribed to this page/comment.
	query := fmt.Sprintf(`
		SELECT userId
		FROM subscriptions %s`, whereClause)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var userId int64
		err := rows.Scan(&userId)
		if err != nil {
			return fmt.Errorf("failed to scan for subscriptions: %v", err)
		}
		if userId == task.UserId {
			return nil
		}

		// Check if we already have a similar update.
		var previousUpdateId int64
		var exists bool
		newCountValue := 1
		query = fmt.Sprintf(`
			SELECT id
			FROM updates
			WHERE type="%s" AND newCount>0 AND
				groupByPageId=%d AND groupByUserId=%d AND
				subscribedToPageId=%d AND subscribedToUserId=%d AND
				goToPageId=%d
			ORDER BY createdAt DESC
			LIMIT 1`,
			task.UpdateType,
			task.GroupByPageId, task.GroupByUserId,
			task.SubscribedToPageId, task.SubscribedToUserId,
			task.GoToPageId)
		exists, err = database.QueryRowSql(c, query, &previousUpdateId)
		if err != nil {
			return fmt.Errorf("failed to check for existing update: %v", err)
		}
		if exists {
			// This is a similar update, so don't count it
			newCountValue = 0
		}

		// Insert new update / update newCount on existing one
		hashmap := make(map[string]interface{})
		if previousUpdateId > 0 {
			hashmap["id"] = previousUpdateId
		}
		hashmap["userId"] = userId
		hashmap["type"] = task.UpdateType
		hashmap["groupByPageId"] = task.GroupByPageId
		hashmap["groupByUserId"] = task.GroupByUserId
		hashmap["subscribedToPageId"] = task.SubscribedToPageId
		hashmap["subscribedToUserId"] = task.SubscribedToUserId
		hashmap["goToPageId"] = task.GoToPageId
		hashmap["createdAt"] = database.Now()
		hashmap["newCount"] = newCountValue
		query = database.GetInsertSql("updates", hashmap, "newCount")
		if _, err = database.ExecuteSql(c, query); err != nil {
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
