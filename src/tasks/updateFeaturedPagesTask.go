// updateFeaturedPagesTask.go checks if the given marks have an answer.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	updateFeaturedPagesPeriod = 1 * 60 * 60 // 1 hour
)

// UpdateFeaturedPagesTask is the object that's put into the daemon queue.
type UpdateFeaturedPagesTask struct {
}

func (task UpdateFeaturedPagesTask) Tag() string {
	return "updateFeaturedPages"
}

// Check if this task is valid, and we can safely execute it.
func (task UpdateFeaturedPagesTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task UpdateFeaturedPagesTask) Execute(db *database.DB) (delay int, err error) {
	delay = updateFeaturedPagesPeriod
	c := db.C

	if err := task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== UPDATE FEATURED PAGES START ====")
	defer c.Infof("==== UPDATE FEATURED PAGES COMPLETED ====")

	// Which pages should be featured
	featuredPageIds := make([]string, 0)

	// Load all pages that haven't been featured yet
	rows := database.NewQuery(`
		SELECT pi.pageId
		FROM`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
		JOIN pageDomainPairs AS pdp /*Has to be part of a domain*/
		ON (pi.pageId=pdp.pageId)
		LEFT JOIN pagePairs AS pp
		ON (pi.pageId=pp.childId)
		WHERE pi.seeGroupId="" AND pi.featuredAt=0 AND pi.type!=?`, core.CommentPageType).Add(`
			AND pp.type=?`, core.TagPagePairType).Add(`
			AND pp.parentId IN (?,?)`, core.AClassPageID, core.BClassPageID).Add(`
		GROUP BY 1`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		if err := rows.Scan(&pageID); err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		featuredPageIds = append(featuredPageIds, pageID)
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("Failed to load featured page candidates: %v", err)
	}

	if len(featuredPageIds) <= 0 {
		return
	}
	c.Infof("New featured pages: %+v", featuredPageIds)

	// Update the database
	statement := database.NewQuery(`
		UPDATE pageInfos
		SET featuredAt=NOW()
		WHERE pageId IN`).AddArgsGroupStr(featuredPageIds).ToStatement(db)
	if _, err = statement.Exec(); err != nil {
		return 0, fmt.Errorf("Failed to update pageInfos: %v", err)
	}

	return
}
