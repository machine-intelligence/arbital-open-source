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

	// Go to destination. One of these has to be set. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageId string

	// If set the update will shown only to maintainers.
	ForceMaintainersOnly bool

	// Optional. FK into changeLogs table.
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

	if task.UpdateType == core.ChangeLogUpdateType && task.ChangeLogId <= 0 {
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

	// Load seeGroupIds for the groupByPage and goToPage. Used to filter out updates for users who
	// won't have permission to click through to the pages linked in the update.
	var requiredGroupMemberships []string
	rows = database.NewQuery(`
		SELECT DISTINCT seeGroupId
		FROM`).AddPart(core.PageInfosTableWithOptions(nil, &core.PageInfosOptions{Deleted: true})).Add(`AS pi
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
	query = database.NewQuery(`
			SELECT DISTINCT s.userId
			FROM subscriptions AS s
			JOIN pages as p
			ON s.userId = p.creatorId
			WHERE s.toId=? AND p.pageId=?`, task.SubscribedToId, task.SubscribedToId)
	if !task.ForceMaintainersOnly && (task.UpdateType == core.TopLevelCommentUpdateType || task.UpdateType == core.ReplyUpdateType ||
		task.UpdateType == core.NewPageByUserUpdateType || task.UpdateType == core.AtMentionUpdateType ||
		task.UpdateType == core.AddedToGroupUpdateType || task.UpdateType == core.RemovedFromGroupUpdateType ||
		task.UpdateType == core.InviteReceivedUpdateType || task.UpdateType == core.ResolvedMarkUpdateType ||
		task.UpdateType == core.AnsweredMarkUpdateType) {
		// This update can be shown to all users who are subsribed
	} else {
		// This update is only for authors who explicitly opted into maintaining the page
		query = query.Add(`AND s.asMaintainer`)
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

func EnqueueRelationshipUpdates(c sessions.Context, userId string,
	parentId string, childId string, parentChangeLogId int64, childChangeLogId int64) error {
	err := enqueueRelationshipUpdatesInternal(c, userId, parentId, childId, false, childChangeLogId)
	if err != nil {
		return err
	}
	// Note: we can't return an error if the second task fails, since the caller might
	// rollback a transaction
	// TODO: use GAE's function to enqueue multiple tasks at once
	enqueueRelationshipUpdatesInternal(c, userId, parentId, childId, true, parentChangeLogId)
	return nil
}

func enqueueRelationshipUpdatesInternal(c sessions.Context, userId string,
	parentId string, childId string, updateIsForChild bool, changeLogId int64) error {
	var task NewUpdateTask
	task.UserId = userId
	task.ChangeLogId = changeLogId
	task.UpdateType = core.ChangeLogUpdateType
	if updateIsForChild {
		task.GroupByPageId = childId
		task.SubscribedToId = childId
		task.GoToPageId = parentId
	} else {
		task.GroupByPageId = parentId
		task.SubscribedToId = parentId
		task.GoToPageId = childId
	}
	return Enqueue(c, &task, nil)
}
