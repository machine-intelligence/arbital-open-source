// propagateDomainTask.go updates all the page's children to have the right domains.
package tasks

import (
	"database/sql"
	"fmt"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
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
func (task *PropagateDomainTask) Execute(c sessions.Context) (delay int, err error) {
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== PROPAGATE DOMAIN START ====")
	defer c.Debugf("==== PROPAGATE DOMAIN COMPLETED SUCCESSFULLY ====")

	// Process the first page.
	// Map of pageId -> whether or not we processed the children
	pageMap := make(map[int64]bool)
	err = propagateDomainToPage(c, task.PageId, pageMap)
	if err != nil {
		c.Debugf("ERROR: %v", err)
		return -1, err
	}
	return 0, nil
}

// propagateDomainToPage forces domain recalculation for the given page.
func propagateDomainToPage(c sessions.Context, pageId int64, pageMap map[int64]bool) error {
	processedChildren, processedPage := pageMap[pageId]
	if !processedPage {
		// Compute what domains the page already has
		domainMap := make(map[int64]*domainFlags)
		query := fmt.Sprintf(`
			SELECT domainId
			FROM pageDomainPairs
			WHERE pageId=%d`, pageId)
		err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
		query = fmt.Sprintf(`
			(SELECT pd.domainId
			FROM pageDomainPairs AS pd
			JOIN pagePairs as pp
			ON (pp.parentId=pd.pageId)
			WHERE childId=%d)
			UNION
			(SELECT id
			FROM groups
			WHERE isDomain AND rootPageId=%d)`, pageId, pageId)
		err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
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
		addDomainIds := make([]string, 0)
		removeDomainIds := make([]string, 0)
		for domainId, flags := range domainMap {
			if flags.ShouldHave && !flags.Has {
				addDomainIds = append(addDomainIds, fmt.Sprintf("(%d,%d)", domainId, pageId))
			} else if !flags.ShouldHave && flags.Has {
				removeDomainIds = append(removeDomainIds, fmt.Sprintf("%d", domainId))
			}
		}

		// Add missing domains
		if len(addDomainIds) > 0 {
			query = fmt.Sprintf(`
				INSERT INTO pageDomainPairs
				(domainId,pageId) VALUES %s`, strings.Join(addDomainIds, ","))
			_, err = database.ExecuteSql(c, query)
			if err != nil {
				return fmt.Errorf("Failed to add to pageDomainPairs: %v", err)
			}
		}

		// Remove obsolete domains
		if len(removeDomainIds) > 0 {
			query = fmt.Sprintf(`
				DELETE FROM pageDomainPairs
				WHERE pageId=%d AND domainId IN (%s)`, pageId, strings.Join(removeDomainIds, ","))
			_, err = database.ExecuteSql(c, query)
			if err != nil {
				return fmt.Errorf("Failed to remove pageDomainPairs: %v", err)
			}
		}

		// Make the page as processed
		// Mark children as processed iff there were no changes
		pageMap[pageId] = len(addDomainIds) <= 0 && len(removeDomainIds) <= 0
	}

	if !processedChildren {
		// Get all the children and add them for processing
		query := fmt.Sprintf(`
			SELECT childId
			FROM pagePairs
			WHERE parentId=%d`, pageId)
		err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
			var childId int64
			if err := rows.Scan(&childId); err != nil {
				return fmt.Errorf("failed to scan for childId: %v", err)
			}
			err := propagateDomainToPage(c, childId, pageMap)
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
