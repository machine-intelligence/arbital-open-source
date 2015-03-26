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
	UserId        int64 // user who performed an action, e.g. creating a comment
	ContextPageId int64 // if appropriate, this is the page relevant to the update
	UpdateType    string

	// One of the following has to be set. Users subscribed to this 'thing' will
	// be notified.
	ToPageId    int64
	ToCommentId int64
	ToUserId    int64
}

// Check if this task is valid, and we can safely execute it.
func (task *NewUpdateTask) IsValid() error {
	if task.UserId <= 0 {
		return fmt.Errorf("User id has to be set")
	} else if task.UpdateType == "" {
		return fmt.Errorf("Update type has to be set")
	} else if task.ToPageId <= 0 && task.ToCommentId <= 0 && task.ToUserId <= 0 {
		return fmt.Errorf("At least on To___Id has to be specified")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *NewUpdateTask) Execute(c sessions.Context) (delay int, err error) {
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	// Figure out the subscriptions query constraint.
	var whereClause string
	if task.ToPageId > 0 {
		whereClause = fmt.Sprintf("WHERE toPageId=%d", task.ToPageId)
	} else if task.ToCommentId > 0 {
		whereClause = fmt.Sprintf("WHERE toCommentId=%d", task.ToCommentId)
	} else if task.ToUserId > 0 {
		whereClause = fmt.Sprintf("WHERE toUserId=%d", task.ToUserId)
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
		var updateId int64
		var exists bool
		query = fmt.Sprintf(`
			SELECT id
			FROM updates
			WHERE userId=%d AND type="%s" AND seen=0 AND contextPageId=%d AND
				fromPageId=%d AND fromCommentId=%d AND fromUserId=%d
			ORDER BY updatedAt DESC
			LIMIT 1`,
			userId, task.UpdateType, task.ContextPageId, task.ToPageId, task.ToCommentId, task.ToUserId)
		exists, err = database.QueryRowSql(c, query, &updateId)
		if err != nil {
			return fmt.Errorf("failed to check for existing update: %v", err)
		}
		if exists {
			// Increase count on an existing update.
			query = fmt.Sprintf(`UPDATE updates SET count=count+1,updatedAt=NOW() WHERE id=%d`, updateId)
			if _, err = database.ExecuteSql(c, query); err != nil {
				c.Inc("update_update_fail")
				c.Errorf("Couldn't create update count on an update: %v", err)
			}
		} else {
			// Insert new update.
			hashmap := make(map[string]interface{})
			hashmap["userId"] = userId
			hashmap["type"] = task.UpdateType
			hashmap["contextPageId"] = task.ContextPageId
			hashmap["fromPageId"] = task.ToPageId
			hashmap["fromCommentId"] = task.ToCommentId
			hashmap["fromUserId"] = task.ToUserId
			hashmap["createdAt"] = database.Now()
			hashmap["updatedAt"] = database.Now()
			hashmap["count"] = 1
			query = database.GetInsertSql("updates", hashmap)
			if _, err = database.ExecuteSql(c, query); err != nil {
				c.Inc("new_update_fail")
				c.Errorf("Couldn't create new update: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		c.Errorf("Couldn't process subscriptions: %v", err)
		return -1, err
	}
	return 0, nil
}
