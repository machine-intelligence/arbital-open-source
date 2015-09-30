// propagateDomainTask.go updates all the page's children to have the right domains.
package tasks

import (
	"fmt"

	"zanaduu3/src/database"
)

// PropagateDomainTask is the object that's put into the daemon queue.
type PropagateDomainTask struct {
	PageId int64
}

type domainFlags struct {
	Has        bool
	ShouldHave bool
}

// Check if this task is valid, and we can safely execute it.
func (task *PropagateDomainTask) IsValid() error {
	if task.PageId <= 0 {
		return fmt.Errorf("PageId needs to be set")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *PropagateDomainTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== PROPAGATE DOMAIN START ====")
	defer c.Debugf("==== PROPAGATE DOMAIN COMPLETED SUCCESSFULLY ====")

	// Process the first page.
	// Map of pageId -> whether or not we processed the children
	pageMap := make(map[int64]bool)
	err = propagateDomainToPage(db, task.PageId, pageMap)
	if err != nil {
		c.Debugf("ERROR: %v", err)
		return -1, err
	}
	return 0, nil
}

// propagateDomainToPage forces domain recalculation for the given page.
func propagateDomainToPage(db *database.DB, pageId int64, pageMap map[int64]bool) error {
	processedChildren, processedPage := pageMap[pageId]
	if !processedPage {
		// Compute what domains the page already has
		domainMap := make(map[int64]*domainFlags)
		rows := db.NewStatement(`
			SELECT domainId
			FROM pageDomainPairs
			WHERE pageId=?`).Query(pageId)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var domainId int64
			if err := rows.Scan(&domainId); err != nil {
				return fmt.Errorf("failed to scan for pageDomainPair: %v", err)
			}
			domainMap[domainId] = &domainFlags{Has: true}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Faled to load pageDomainPairs: %v", err)
		}

		// Compute what domains the page should have based on its parents and
		// whether or not it's a root page for some domain
		rows = db.NewStatement(`
			(SELECT pd.domainId
			FROM pageDomainPairs AS pd
			JOIN pagePairs as pp
			ON (pp.parentId=pd.pageId)
			WHERE childId=?)
			UNION
			(SELECT id
			FROM groups
			WHERE isDomain AND rootPageId=?)`).Query(pageId, pageId)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var domainId int64
			if err := rows.Scan(&domainId); err != nil {
				return fmt.Errorf("failed to scan for pageDomainPair: %v", err)
			}
			if flags, ok := domainMap[domainId]; ok {
				flags.ShouldHave = true
			} else {
				domainMap[domainId] = &domainFlags{ShouldHave: true}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Faled to load pageDomainPairs2: %v", err)
		}

		// Compute which domains to add/remove
		addDomainArgs := make([]interface{}, 0)
		removeDomainArgs := make([]interface{}, 0)
		for domainId, flags := range domainMap {
			if flags.ShouldHave && !flags.Has {
				addDomainArgs = append(addDomainArgs, domainId, pageId)
			} else if !flags.ShouldHave && flags.Has {
				removeDomainArgs = append(removeDomainArgs, domainId)
			}
		}

		// Add missing domains
		if len(addDomainArgs) > 0 {
			statement := db.NewStatement(`
				INSERT INTO pageDomainPairs
				(domainId,pageId) VALUES ` + database.ArgsPlaceholder(len(addDomainArgs), 2))
			if _, err = statement.Exec(addDomainArgs...); err != nil {
				return fmt.Errorf("Failed to add to pageDomainPairs: %v", err)
			}
		}

		// Remove obsolete domains
		if len(removeDomainArgs) > 0 {
			statement := db.NewStatement(`
				DELETE FROM pageDomainPairs
				WHERE pageId=? AND domainId IN ` + database.InArgsPlaceholder(len(removeDomainArgs)))
			args := append([]interface{}{pageId}, removeDomainArgs...)
			if _, err = statement.Exec(args...); err != nil {
				return fmt.Errorf("Failed to remove pageDomainPairs: %v", err)
			}
		}

		// Make the page as processed
		// Mark children as processed iff there were no changes
		pageMap[pageId] = len(addDomainArgs) <= 0 && len(removeDomainArgs) <= 0
	}

	if !processedChildren {
		// Get all the children and add them for processing
		rows := db.NewStatement(`
			SELECT childId
			FROM pagePairs
			WHERE parentId=?`).Query(pageId)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var childId int64
			if err := rows.Scan(&childId); err != nil {
				return fmt.Errorf("failed to scan for childId: %v", err)
			}
			err := propagateDomainToPage(db, childId, pageMap)
			return err
		})
		if err != nil {
			return err
		}

		// Mark the page's children processed
		pageMap[pageId] = true
	}

	return nil
}
