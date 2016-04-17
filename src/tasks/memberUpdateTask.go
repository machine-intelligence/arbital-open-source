// grouptionUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// MemberUpdateTask is the object that's put into the daemon queue.
type MemberUpdateTask struct {
	// User who performed the action
	UserId     string
	UpdateType string

	// Member is added to/removed from the given group
	MemberId string
	GroupId  string
}

func (task MemberUpdateTask) Tag() string {
	return "memberUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task MemberUpdateTask) IsValid() error {
	if !core.IsIdValid(task.UserId) {
		return fmt.Errorf("UserId has to be set")
	} else if !core.IsIdValid(task.MemberId) {
		return fmt.Errorf("MemberId has to be set")
	} else if task.UpdateType != core.AddedToGroupUpdateType &&
		task.UpdateType != core.RemovedFromGroupUpdateType {
		return fmt.Errorf("Update type is incorrect")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task MemberUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		c.Errorf("Invalid group update task: %s", err)
		return -1, err
	}

	// Insert new update
	hashmap := make(map[string]interface{})
	hashmap["userId"] = task.MemberId
	hashmap["byUserId"] = task.UserId
	hashmap["type"] = task.UpdateType
	hashmap["groupByPageId"] = task.GroupId
	hashmap["goToPageId"] = task.GroupId
	hashmap["createdAt"] = database.Now()
	hashmap["unseen"] = true
	statement := db.NewInsertStatement("updates", hashmap)
	if _, err = statement.Exec(); err != nil {
		return -1, fmt.Errorf("Couldn't create new update: %v", err)
	}
	return 0, nil
}
