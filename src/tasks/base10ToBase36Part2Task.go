// base10ToBase36Part2Task.go does part 2 of converting all the ids from base 10 to base 36
package tasks

import (
	"fmt"

	"zanaduu3/src/database"
)

// Base10ToBase36Part2Task is the object that's put into the daemon queue.
type Base10ToBase36Part2Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part2Task) IsValid() error {
	return nil
}

var tableName string
var columnName string

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part2Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 2 START ====")
	defer c.Debugf("==== PART 2 COMPLETED ====")

	replaceBatch(db, "pages", "pageId")
	replaceBatch(db, "pages", "creatorId")

	replaceBatch(db, "changeLogs", "userId")
	replaceBatch(db, "changeLogs", "pageId")
	replaceBatch(db, "changeLogs", "auxPageId")

	replaceBatch(db, "groupMembers", "userId")
	replaceBatch(db, "groupMembers", "groupId")

	replaceBatch(db, "likes", "userId")
	replaceBatch(db, "likes", "pageId")

	replaceBatch(db, "links", "parentId")

	replaceBatch(db, "pageDomainPairs", "pageId")
	replaceBatch(db, "pageDomainPairs", "domainId")

	replaceBatch(db, "pageInfos", "pageId")
	replaceBatch(db, "pageInfos", "lockedBy")
	replaceBatch(db, "pageInfos", "seeGroupId")
	replaceBatch(db, "pageInfos", "editGroupId")
	replaceBatch(db, "pageInfos", "createdBy")
	replaceBatch(db, "pageInfos", "alias")

	replaceBatch(db, "pagePairs", "parentId")
	replaceBatch(db, "pagePairs", "childId")

	replaceBatch(db, "pageSummaries", "pageId")

	replaceBatch(db, "subscriptions", "userId")
	replaceBatch(db, "subscriptions", "toId")

	replaceBatch(db, "updates", "userId")
	replaceBatch(db, "updates", "groupByPageId")
	replaceBatch(db, "updates", "groupByUserId")
	replaceBatch(db, "updates", "subscribedToId")
	replaceBatch(db, "updates", "goToPageId")
	replaceBatch(db, "updates", "byUserId")

	replaceBatch(db, "userMasteryPairs", "userId")
	replaceBatch(db, "userMasteryPairs", "masteryId")

	replaceBatch(db, "users", "id")

	replaceBatch(db, "visits", "userId")
	replaceBatch(db, "visits", "pageId")

	replaceBatch(db, "votes", "userId")
	replaceBatch(db, "votes", "pageId")

	doOneQuery(db, `UPDATE pageInfos SET aliasBase36 = alias WHERE aliasBase36 = "";`)

	return 0, err
}

func replaceBatch(db *database.DB, newTableName string, newColumnName string) error {
	tableName = newTableName
	columnName = newColumnName

	statement := db.NewStatement(`UPDATE ` + tableName + ` SET ` + columnName + `Processed = 1, ` + columnName + `Base36 = (SELECT base36Id from base10tobase36 WHERE base10Id = ` + tableName + `.` + columnName + `) WHERE ` + columnName + `Processed = 0`)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update table "+tableName+", column "+columnName+": %v", err)
	}
	statement.Close()

	return nil
}
