// copyPagesTask.go takes some pages and makes a copy of them.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// CopyPagesTask is the object that's put into the daemon queue.
type CopyPagesTask struct {
}

func (task CopyPagesTask) Tag() string {
	return "copyPages"
}

// Check if this task is valid, and we can safely execute it.
func (task CopyPagesTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task CopyPagesTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err := task.IsValid(); err != nil {
		return 0, err
	}

	c.Infof("==== COPY PAGES START ====")
	defer c.Infof("==== COPY PAGES COMPLETED ====")

	// Which pages should be featured
	// Load all pages that haven't been featured yet
	rows := database.NewQuery(`
		select p2.pageId,p2.edit from (
			select p.pageId,p.edit
			from pages as p
			join pageInfos as pi
			on (p.pageId=pi.pageId)
			join pageDomainPairs as pdp
			on (pi.pageId=pdp.pageId)
			where p.creatorId="2" and p.createdAt < "2016-04-03"
				and not isautosave and not issnapshot
				and pi.type!="comment" and pdp.domainId="1lw"
			order by p.createdAt desc
		) as p2
		group by p2.pageId`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		var editNum int
		if err := rows.Scan(&pageID, &editNum); err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		var newPageID string
		err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
			newPageID, err = core.GetNextAvailableID(tx)
			if err != nil {
				return sessions.NewError("Couldn't get next available Id", err)
			}
			return nil
		})
		if err2 != nil {
			return fmt.Errorf("Failed to get id: %v", err)
		}

		c.Infof("============= (%v,%v) => %v", pageID, editNum, newPageID)

		// Copy pageInfo row
		_, err := database.NewQuery(`
			CREATE TABLE tmptable_1 SELECT * FROM pageInfos WHERE pageId = ?`, pageID).Add(`
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pageInfos row: %v", err)
		}

		_, err = database.NewQuery(`
			UPDATE tmptable_1 SET pageId=?,alias=?,seeGroupId=?,currentEdit=?`, newPageID, newPageID, "4f", editNum).Add(`
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pageInfos row: %v", err)
		}

		_, err = database.NewQuery(`
			INSERT INTO pageInfos SELECT * FROM tmptable_1
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pageInfos row: %v", err)
		}

		_, err = database.NewQuery(`
			DROP TABLE IF EXISTS tmptable_1;
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to delete temp pageInfos table: %v", err)
		}

		_, err = database.NewQuery(`
			INSERT INTO pagePairs (parentId,childId,type,creatorId,createdAt) VALUES
					(?,?,"parent",?,now())`, "4f", newPageID, "4f").Add(`
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to delete temp pageInfos table: %v", err)
		}

		// Copy pages row
		_, err = database.NewQuery(`
			CREATE TABLE tmptable_1 SELECT * FROM pages WHERE pageId=? AND edit=?;`, pageID, editNum).Add(`
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pages row: %v", err)
		}
		_, err = database.NewQuery(`
			UPDATE tmptable_1 SET pageId=?,isLiveEdit=true;`, newPageID).Add(`
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pages row: %v", err)
		}
		_, err = database.NewQuery(`
			INSERT INTO pages SELECT * FROM tmptable_1;
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pages row: %v", err)
		}
		_, err = database.NewQuery(`
			DROP TABLE IF EXISTS tmptable_1;
		`).ToStatement(db).Exec()
		if err != nil {
			return fmt.Errorf("Failed to copy pages row: %v", err)
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("Failed to load pages: %v", err)
	}

	return
}
