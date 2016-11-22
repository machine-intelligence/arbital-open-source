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
	UserID     string
	UpdateType string

	// We'll notify the users who are subscribed to this page id (also could be a
	// user id, group id, domain id)
	SubscribedToID string

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageID string

	// If set the update will shown only to maintainers.
	ForceMaintainersOnly bool

	// Optional. FK into changeLogs table.
	ChangeLogID int64

	// Only set if UpdateType is for a mark. Id is a FK into marks table.
	MarkID string
}

func (task NewUpdateTask) Tag() string {
	return "newUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task NewUpdateTask) IsValid() error {
	if !core.IsIDValid(task.UserID) {
		return fmt.Errorf("User id has to be set: %v", task.UserID)
	} else if task.UpdateType == "" {
		return fmt.Errorf("Update type has to be set")
	} else if !core.IsIDValid(task.SubscribedToID) {
		return fmt.Errorf("SubscibedTo id has to be set")
	}

	if !core.IsIDValid(task.GoToPageID) {
		return fmt.Errorf("GoToPageId has to be set")
	}

	if task.UpdateType == core.ChangeLogUpdateType && task.ChangeLogID <= 0 {
		return fmt.Errorf("No changeLogId set for a ChangeLogUpdateType")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task NewUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C
	var rows *database.Rows

	if err = task.IsValid(); err != nil {
		return -1, fmt.Errorf("Invalid new update task: %v", err)
	}

	var query *database.QueryPart
	// Iterate through all users who are subscribed to this page/comment.
	// If it is an editors only comment, only select editor ids.
	query = database.NewQuery(`
			SELECT DISTINCT s.userId
			FROM subscriptions AS s
			JOIN pages as p
			ON (s.userId = p.creatorId)
			JOIN pageInfos AS pi
			ON (p.pageId = pi.pageId)
			WHERE s.toId=?`, task.SubscribedToID).Add(`
				AND p.pageId=?`, task.SubscribedToID).Add(`
				AND pi.seeDomainId != 0`)
	if !task.ForceMaintainersOnly &&
		(task.UpdateType == core.TopLevelCommentUpdateType || task.UpdateType == core.ReplyUpdateType ||
			task.UpdateType == core.NewPageByUserUpdateType || task.UpdateType == core.AtMentionUpdateType ||
			task.UpdateType == core.AddedToGroupUpdateType || task.UpdateType == core.RemovedFromGroupUpdateType ||
			task.UpdateType == core.InviteReceivedUpdateType || task.UpdateType == core.ResolvedMarkUpdateType ||
			task.UpdateType == core.AnsweredMarkUpdateType) {
		// This update can be shown to all users who are subscribed
	} else {
		// This update is only for authors who explicitly opted into maintaining the page
		query = query.Add(`AND s.asMaintainer`)
	}
	rows = query.ToStatement(db).Query()
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
		hashmap := make(database.InsertMap)
		hashmap["userId"] = userID
		hashmap["byUserId"] = task.UserID
		hashmap["type"] = task.UpdateType
		hashmap["subscribedToId"] = task.SubscribedToID
		hashmap["goToPageId"] = task.GoToPageID
		hashmap["changeLogId"] = task.ChangeLogID
		hashmap["markId"] = task.MarkID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("updates", hashmap)
		if _, err = statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't create new update: %v", err)
		}
		return nil
	})
	if err != nil {
		c.Inc("new_update_fail")
		return -1, fmt.Errorf("Couldn't process subscriptions: %v", err)
	}
	return 0, nil
}
