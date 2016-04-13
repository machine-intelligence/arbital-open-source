// updateElasticPageTask.go adds all the pages to the elastic index.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// UpdateElasticPageTask is the object that's put into the daemon queue.
type UpdateElasticPageTask struct {
	PageId string
}

// Check if this task is valid, and we can safely execute it.
func (task *UpdateElasticPageTask) IsValid() error {
	if !core.IsIdValid(task.PageId) {
		return fmt.Errorf("Invalid page id: %s", task.PageId)
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *UpdateElasticPageTask) Execute(db *database.DB) (int, error) {
	c := db.C

	if err := task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("Updaing elastic page: %s", task.PageId)

	// Compute all priors.
	rows := db.NewStatement(`
		SELECT p.pageId,pi.type,p.title,p.clickbait,p.text,pi.alias,pi.seeGroupId,pi.createdBy
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		WHERE isLiveEdit AND p.pageId=?`).Query(task.PageId)
	err := rows.Process(populateElasticProcessPage)
	if err != nil {
		return -1, err
	}
	return 0, err
}
