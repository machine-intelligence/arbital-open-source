// updateElasticPageTask.go adds all the pages to the elastic index.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// UpdateElasticPageTask is the object that's put into the daemon queue.
type UpdateElasticPageTask struct {
	PageID string
}

func (task UpdateElasticPageTask) Tag() string {
	return "updateElasticPage"
}

// Check if this task is valid, and we can safely execute it.
func (task UpdateElasticPageTask) IsValid() error {
	if !core.IsIDValid(task.PageID) {
		return fmt.Errorf("Invalid page id: %s", task.PageID)
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task UpdateElasticPageTask) Execute(db *database.DB) (int, error) {
	c := db.C

	if err := task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("Updaing elastic page: %s", task.PageID)

	// Compute all priors.
	rows := database.NewQuery(`
		SELECT p.pageId,pi.type,p.title,p.clickbait,p.text,pi.alias,pi.seeDomainId,pi.createdBy,pi.externalUrl,pi.hasVote
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		WHERE p.isLiveEdit
			AND p.pageId=?`, task.PageID).Add(`
			AND`).AddPart(core.PageInfosFilter(nil)).ToStatement(db).Query()
	err := rows.Process(populateElasticProcessPage)
	if err != nil {
		return -1, err
	}
	return 0, err
}
