// propagateDomainTask.go updates all the page's children to have the right domains.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// PropagateDomainTask is the object that's put into the daemon queue.
type PropagateDomainTask struct {
	PageId string
	// If true, the page was deleted and we should update children + parents
	Deleted bool
}

func (task PropagateDomainTask) Tag() string {
	return "propagateDomain"
}

// Check if this task is valid, and we can safely execute it.
func (task PropagateDomainTask) IsValid() error {
	if !core.IsIdValid(task.PageId) {
		return fmt.Errorf("PageId needs to be set")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task PropagateDomainTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== PROPAGATE DOMAIN START ====")
	defer c.Infof("==== PROPAGATE DOMAIN COMPLETED ====")

	err = propagateDomainsToPageAndDescendants(db, task.PageId)
	if err != nil {
		return -1, fmt.Errorf("Error propagating domain: %v", err)
	}

	return 0, nil
}

func propagateDomainsToPageAndDescendants(db *database.DB, pageId string) error {
	// all the descendants of the page (plus the page itself)
	pagesToUpdate, err := core.GetDescendants(db, pageId)
	if err != nil {
		return err
	}

	return core.PropagateDomains(db, pagesToUpdate)
}
