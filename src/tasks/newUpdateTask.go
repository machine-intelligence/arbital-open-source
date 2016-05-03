// newUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
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

	// If it is an editors only comment, only notify editors
	EditorsOnly bool

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageId string

	// Only set if UpdateType is 'pageInfoEdit'. Id is a FK into changeLogs table.
	ChangeLogId int64

	// Only set if UpdateType is for a mark. Id is a FK into marks table.
	MarkId string
}

func (task NewUpdateTask) Tag() string {
	return "newUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task NewUpdateTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("User id has to be set: %v", task.UserId)
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
func (task NewUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C
	var rows *database.Rows

	if err = task.IsValid(); err != nil {
		return -1, fmt.Errorf("Invalid new update task: %v", err)
	}

	// Load seeGroupIds for the groupByPage and goToPage. Used to filter out updates for users who
	// won't have permission to click through to the pages linked in the update.
	var requiredGroupMemberships []string
	rows = database.NewQuery(`
		SELECT DISTINCT seeGroupId
		FROM pageInfos
		WHERE seeGroupId != '' AND pageId IN (?,?)`, task.GroupByPageId, task.GoToPageId).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var groupId string
		err := rows.Scan(&groupId)
		if err != nil {
			return fmt.Errorf("failed to scan for required groups: %v", err)
		}

		requiredGroupMemberships = append(requiredGroupMemberships, groupId)
		return nil
	})
	if err != nil {
		return -1, fmt.Errorf("Couldn't process group requirements: %v", err)
	}

	var query *database.QueryPart
	// Iterate through all users who are subscribed to this page/comment.
	// If it is an editors only comment, only select editor ids.
	if task.EditorsOnly {
		query = database.NewQuery(`
			SELECT DISTINCT s.userId
			FROM subscriptions AS s
			JOIN pages as p
			ON s.userId = p.creatorId
			WHERE s.toId=? AND p.pageId=?`, task.SubscribedToId, task.SubscribedToId)
	} else {
		query = database.NewQuery(`
			SELECT s.userId
			FROM subscriptions AS s
			WHERE s.toId=?`, task.SubscribedToId)
	}
	if len(requiredGroupMemberships) > 0 {
		query = query.Add(`AND
		(
			SELECT COUNT(*)
			FROM groupMembers AS gm
			WHERE gm.userId = s.userId AND gm.groupId IN`).AddArgsGroupStr(requiredGroupMemberships).Add(`
		) = ?`, len(requiredGroupMemberships))
	}
	rows = query.ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userId string
		err := rows.Scan(&userId)
		if err != nil {
			return fmt.Errorf("failed to scan for subscriptions: %v", err)
		}
		if userId == task.UserId {
			return nil
		}

		// Insert new update
		hashmap := make(database.InsertMap)
		hashmap["userId"] = userId
		hashmap["byUserId"] = task.UserId
		hashmap["type"] = task.UpdateType
		hashmap["groupByPageId"] = task.GroupByPageId
		hashmap["groupByUserId"] = task.GroupByUserId
		hashmap["subscribedToId"] = task.SubscribedToId
		hashmap["goToPageId"] = task.GoToPageId
		hashmap["changeLogId"] = task.ChangeLogId
		hashmap["markId"] = task.MarkId
		hashmap["createdAt"] = database.Now()
		hashmap["unseen"] = true
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

func EnqueueNewRelationshipUpdates(c sessions.Context, userId string, pairType string, childPageType string,
	parentId string, childId string) {
	enqueueRelationshipUpdatesInternal(c, userId, pairType, childPageType, parentId, childId, false, false)
	enqueueRelationshipUpdatesInternal(c, userId, pairType, childPageType, parentId, childId, true, false)
}

func EnqueueDeleteRelationshipUpdates(c sessions.Context, userId string, pairType string, childPageType string,
	parentId string, childId string) {
	enqueueRelationshipUpdatesInternal(c, userId, pairType, childPageType, parentId, childId, false, true)
	enqueueRelationshipUpdatesInternal(c, userId, pairType, childPageType, parentId, childId, true, true)
}

func enqueueRelationshipUpdatesInternal(c sessions.Context, userId string, pairType string, childPageType string,
	parentId string, childId string, updateIsForChild bool, relationshipIsDeleted bool) {

	var task NewUpdateTask
	task.UserId = userId
	updateType, err := core.GetUpdateTypeForPagePair(pairType, childPageType, updateIsForChild, relationshipIsDeleted)
	if err != nil {
		c.Errorf("Couldn't get the update type for a page pair type: %v", err)
	}

	task.UpdateType = updateType
	if updateIsForChild {
		task.GroupByPageId = childId
		task.SubscribedToId = childId
		task.GoToPageId = parentId
	} else {
		task.GroupByPageId = parentId
		task.SubscribedToId = parentId
		task.GoToPageId = childId
	}
	if err := Enqueue(c, &task, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
