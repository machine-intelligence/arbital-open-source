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

type domainFlags struct {
	Has        bool
	ShouldHave bool
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

	// Process the first page.
	// Map of pageId -> whether or not we processed the children
	pageMap := make(map[string]bool)
	err = propagateDomainToPage(db, task.PageId, pageMap)
	if err != nil {
		return -1, fmt.Errorf("Error propagating domain: %v", err)
	}

	return 0, nil
}

// propagateDomainToPage forces domain recalculation for the given page.
// It uses a depth-first search to process the children of the page.
func propagateDomainToPage(db *database.DB, pageId string, pageMap map[string]bool) error {
	db.C.Infof("Processing %s", pageId)
	processedChildren, processedPage := pageMap[pageId]
	if !processedPage {
		// Compute what domains the page already has
		domainMap := make(map[string]*domainFlags)
		rows := db.NewStatement(`
			SELECT domainId
			FROM pageDomainPairs
			WHERE pageId=?`).Query(pageId)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var domainId string
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
		rows = database.NewQuery(`
			(
				/* get the domains (and see-groups) of parents */
				SELECT pdp.domainId, pi.seeGroupId
				FROM (
					SELECT parentId AS pageId
					FROM pagePairs AS pp
					WHERE pp.childId=?`, pageId).Add(`AND pp.type=?`, core.ParentPagePairType).Add(`
				) AS parents
				JOIN pageDomainPairs AS pdp ON pdp.pageId=parents.pageId
				JOIN`).AddPart(core.PageInfosTable(nil)).Add(`AS pi ON pi.pageId=parents.pageId
			) UNION (
				/* include the given page, if it's the root of a domain */
				SELECT pi.pageId, pi.seeGroupId
				FROM`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
				WHERE pi.pageId=?`, pageId).Add(`AND pi.type=?`, core.DomainPageType).Add(`
			)`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var domainId, seeGroupId string
			if err := rows.Scan(&domainId, &seeGroupId); err != nil {
				return fmt.Errorf("failed to scan for pageDomainPair: %v", err)
			}
			if flags, ok := domainMap[domainId]; ok {
				flags.ShouldHave = !core.IsIdValid(seeGroupId)
			} else {
				domainMap[domainId] = &domainFlags{ShouldHave: seeGroupId == ""}
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
			WHERE parentId=? AND type=?`).Query(pageId, core.ParentPagePairType)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var childId string
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
