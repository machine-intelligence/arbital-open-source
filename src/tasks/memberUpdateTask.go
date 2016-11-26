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
	UserID     string
	UpdateType string

	// Member is added to/removed from the given group
	MemberID string
	DomainID string
}

func (task MemberUpdateTask) Tag() string {
	return "memberUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task MemberUpdateTask) IsValid() error {
	if !core.IsIDValid(task.UserID) {
		return fmt.Errorf("UserId has to be set")
	} else if !core.IsIDValid(task.MemberID) {
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
		c.Errorf("Invalid group update task: %v", err)
		return -1, err
	}

	d, err := core.LoadDomainByID(db, task.DomainID)
	if err != nil {
		c.Errorf("Couldn't load domain: %v", err)
		return -1, err
	}

	// Insert new update
	hashmap := make(map[string]interface{})
	hashmap["userId"] = task.MemberID
	hashmap["byUserId"] = task.UserID
	hashmap["type"] = task.UpdateType
	hashmap["goToPageId"] = d.PageID
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("updates", hashmap)
	if _, err = statement.Exec(); err != nil {
		return -1, fmt.Errorf("Couldn't create new update: %v", err)
	}
	return 0, nil
}
