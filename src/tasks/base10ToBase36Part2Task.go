// base10ToBase36Part2Task.go does part 2 of converting all the ids from base 10 to base 36
package tasks

import (
	"fmt"
	//"regexp"
	//"strings"

	"zanaduu3/src/database"
	//"zanaduu3/src/user"
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
	replaceBatch(db, "pages", "privacyKey")

	replaceBatch(db, "changeLogs", "pageId")
	replaceBatch(db, "changeLogs", "auxPageId")

	replaceBatch(db, "fixedIds", "pageId")

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

	return 0, err
}

func replaceBatch(db *database.DB, newTableName string, newColumnName string) error {
	tableName = newTableName
	columnName = newColumnName

	//rows := db.NewStatement(`SELECT ` + columnName + ` FROM ` + tableName + ` WHERE 1`).Query()
	rows := db.NewStatement(`SELECT ` + columnName + ` FROM ` + tableName + ` WHERE ` + columnName + `Processed = 0`).Query()

	if err := rows.Process(replaceId); err != nil {
		db.C.Debugf("ERROR, failed to replace batch table "+tableName+", column "+columnName+": %v", err)
		return err
	}
	return nil
}

func replaceId(db *database.DB, rows *database.Rows) error {
	var base10id string
	if err := rows.Scan(&base10id); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	//db.C.Debugf("base10id: %v", base10id)

	var base36id string
	row := db.NewStatement(`SELECT base36id FROM base10tobase36 WHERE base10id=?`).QueryRow(base10id)
	//row := db.QueryRow(`SELECT base36id FROM base10tobase36 WHERE base10id=` + base10id)

	_, err := row.Scan(&base36id)
	if err != nil {
		return fmt.Errorf("Error reading base36id: %v", err)
	}

	//db.C.Debugf("base36id: %v", base36id)

	hashmap := make(map[string]interface{})
	hashmap[columnName] = base36id

	//queryString := `UPDATE ` + tableName + ` SET ` + columnName + ` = "` + base36id + `" WHERE ` + columnName + ` = "` + base10id + `"`
	queryString := `UPDATE ` + tableName + ` SET ` + columnName + `Base36 = "` + base36id + `", ` + columnName + `Processed = 1 WHERE ` + columnName + ` = "` + base10id + `"`

	//db.C.Debugf("queryString: %v", queryString)
	statement := db.NewStatement(queryString)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update table "+tableName+", column "+columnName+", base10id "+base10id+": %v", err)
	}
	statement.Close()

	return nil
}
