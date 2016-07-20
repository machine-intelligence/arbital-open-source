// updatePagePairsTask.go tries to publish all relationships for the given page.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// UpdatePagePairsTask is the object that's put into the daemon queue.
type UpdatePagePairsTask struct {
	PageID string
}

func (task UpdatePagePairsTask) Tag() string {
	return "updatePagePairs"
}

// Check if this task is valid, and we can safely execute it.
func (task UpdatePagePairsTask) IsValid() error {
	if !core.IsIDValid(task.PageID) {
		return fmt.Errorf("PageId needs to be set")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task UpdatePagePairsTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== UPDATE RELATIONSHIPS START ====")
	defer c.Infof("==== UPDATE RELATIONSHIPS COMPLETED ====")

	// Load relationships which haven't been published yet
	queryPart := database.NewQuery(`
		WHERE (pp.childId=? OR pp.parentId=?)`, task.PageID, task.PageID).Add(`
			AND NOT pp.everPublished`)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pagePair *core.PagePair) error {
		var task PublishPagePairTask
		task.UserID = pagePair.CreatorID
		task.PagePairID = pagePair.ID
		return Enqueue(c, &task, nil)
	})
	if err != nil {
		return -1, fmt.Errorf("Failed to load pending relationships: %v", err)
	}

	return 0, nil
}
