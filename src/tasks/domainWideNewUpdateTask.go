// domainWideNewUpdateTask.go inserts corresponding update.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// DomainWideNewUpdateTask is the object that's put into the daemon queue.
type DomainWideNewUpdateTask struct {
	// User who performed an action, e.g. creating a comment
	UserID     string
	UpdateType string

	// Grouping key. One of these has to set. We'll group all updates by this key
	// to show in one panel.
	// Domain for which this update happens
	DomainID string

	// Go to destination. This is where we'll direct
	// the user if they want to see more info about this update, e.g. to see the
	// comment someone made.
	GoToPageID string
}

func (task DomainWideNewUpdateTask) Tag() string {
	return "domainWideNewUpdate"
}

// Check if this task is valid, and we can safely execute it.
func (task DomainWideNewUpdateTask) IsValid() error {
	if !core.IsIDValid(task.UserID) {
		return fmt.Errorf("User id has to be set: %v", task.UserID)
	} else if task.UpdateType == "" {
		return fmt.Errorf("Update type has to be set")
	} else if !core.IsIDValid(task.DomainID) {
		return fmt.Errorf("Domain id has to be set")
	}

	if !core.IsIDValid(task.GoToPageID) {
		return fmt.Errorf("GoToPageId has to be set")
	}

	if task.UpdateType != core.PageToDomainSubmissionUpdateType {
		return fmt.Errorf("Invalid update type")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task DomainWideNewUpdateTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, fmt.Errorf("Invalid new update task: %v", err)
	}

	reviewerRoles := core.RolesAtLeast(core.ReviewerDomainRole)

	// Iterate through all users who are members of the domain.
	rows := database.NewQuery(`
		SELECT userId
		FROM domainMembers
		WHERE domainId=?`, task.DomainID).Add(`
			AND role IN`).AddArgsGroupStr(reviewerRoles).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}

		// Insert new update
		hashmap := make(database.InsertMap)
		hashmap["userId"] = userID
		hashmap["byUserId"] = task.UserID
		hashmap["type"] = task.UpdateType
		hashmap["subscribedToId"] = task.DomainID
		hashmap["goToPageId"] = task.GoToPageID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("updates", hashmap)
		if _, err = statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't create new update: %v", err)
		}
		return nil
	})
	if err != nil {
		c.Inc("new_update_fail")
		return -1, fmt.Errorf("Couldn't insert all updates: %v", err)
	}
	return 0, nil
}
