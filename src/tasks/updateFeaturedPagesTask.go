// updateFeaturedPagesTask.go checks if the given marks have an answer.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	updateFeaturedPagesPeriod   = 1 * 60 * 60 // 1 hour
	minLengthToBeFeatured       = 2000        // characters
	suppressingTagsParentPageId = "3zb"
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

	// Load which tags suppress a page from being featured
	suppressingTagIds, err := core.LoadMetaTags(db, suppressingTagsParentPageId)
	if err != nil {
		return 0, fmt.Errorf("Couldn't load meta tags: %v", err)
	}

	// Which pages should be featured
	featuredPageIds := make([]string, 0)

	// Load all pages that haven't been featured yet
	rows := database.NewQuery(`
		SELECT pi.pageId
		FROM`).AddPart(core.PageInfosTable(nil)).Add(`AS pi
		JOIN pages AS p
		ON (pi.pageId=p.pageId)
		LEFT JOIN pagePairs AS pp
		ON (pp.childId=pi.pageId)
		WHERE p.isLiveEdit AND length(p.text)>=?`, minLengthToBeFeatured).Add(`
			AND pi.seeGroupId="" AND pi.featuredAt=0 AND pi.type!=?`, core.CommentPageType).Add(`
			AND pp.type=?`, core.TagPagePairType).Add(`
		GROUP BY 1
		HAVING SUM(pp.parentId IN`).AddArgsGroupStr(suppressingTagIds).Add(`) <= 0`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		if err := rows.Scan(&pageId); err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		featuredPageIds = append(featuredPageIds, pageId)
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
