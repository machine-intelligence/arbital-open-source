// fixTextTask.go updates all pages' text fields to fix common mistakes
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

// FixTextTask is the object that's put into the daemon queue.
type FixTextTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *FixTextTask) IsValid() error {
	return nil
}

var lastBase36Id string
var tableName string
var columnName string

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *FixTextTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== FIX TEXT START ====")
	defer c.Debugf("==== FIX TEXT COMPLETED ====")
	/*
		rows := db.NewStatement(`
			SELECT pageId,edit,text
			FROM pages
			WHERE isCurrentEdit`).Query()
		if err = rows.Process(fixText); err != nil {
			c.Debugf("ERROR, failed to fix text: %v", err)
			return 0, err
		}
	*/
	lastBase36Id = "0"
	rows := db.NewStatement(`
			SELECT base10id,createdAt
			FROM base10tobase36
			WHERE 1
			ORDER BY createdAt
			`).Query()
	if err = rows.Process(updateBase10ToBase36); err != nil {
		c.Debugf("ERROR, failed to fix text: %v", err)
		return 0, err
	}

	db.NewStatement(`DELETE FROM base10tobase36 WHERE base36id=""`).Query()

	//replaceBatch(db, "pages", "pageId2")

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

func fixText(db *database.DB, rows *database.Rows) error {
	var pageId, edit string
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}
	/*
		// Find and replace [token1 token2] with [ token1 token2]
		exp := regexp.MustCompile("(\\[[^ \\\\0-9:-\\]]+ [^\\]]*?\\])(?:[^(]|$)")
		newText := exp.ReplaceAllStringFunc(text, func(submatch string) string {
			parts := strings.Split(submatch, " ")
			parts[0] = "[ " + strings.Split(parts[0], "[")[1]
			return strings.Join(parts, " ")
		})
		if newText != text {
			db.C.Debugf("========================== %s", text)
			db.C.Debugf("========================== %s", newText)
			hashmap := make(map[string]interface{})
			hashmap["pageId"] = pageId
			hashmap["edit"] = edit
			hashmap["text"] = newText
			statement := db.NewInsertStatement("pages", hashmap, "text")
			if _, err := statement.Exec(); err != nil {
				return fmt.Errorf("Couldn't update pages table: %v", err)
			}
		}
	*/

	// Find and replace [text](id/alias) links with [id/alias text]

	// First remove all instances of "http://zanaduu3.appspot.com/pages/" in the links, leaving just the pageId
	// On the first pass, accept anything inside the parentheses, since the text we want to remove isn't a valid alias
	exp := regexp.MustCompile("\\[([^\\]]+)\\]\\(([^\\)]+)\\)")
	newText := exp.ReplaceAllStringFunc(text, func(submatch string) string {
		result := submatch
		result = strings.Replace(result, "http://zanaduu3.appspot.com/pages/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/edit/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/pages/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/e/", "", -1)
		//result = strings.Replace(result, "http://arbital.com/p/", "", -1)
		db.C.Debugf("submatch: %v", submatch)
		db.C.Debugf("result  : %v", result)
		return result
	})

	// Now convert from [text](id/alias) to [id/alias text]
	// On this pass, only accept valid aliases inside the parentheses, to prevent changing URL links
	exp = regexp.MustCompile("\\[([^\\]]+)\\]\\(([A-Za-z0-9_]+)\\)")
	newText = exp.ReplaceAllString(newText, "[$2 $1]")

	if newText != text {
		db.C.Debugf("========================== %s", text)
		db.C.Debugf("========================== %s", newText)
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = edit
		hashmap["text"] = newText
		statement := db.NewInsertStatement("pages", hashmap, "text")
		if _, err := statement.Exec(); err != nil {
			return fmt.Errorf("Couldn't update pages table: %v", err)
		}
	}

	return nil
}

func updateBase10ToBase36(db *database.DB, rows *database.Rows) error {
	var base10Id, createdAt string
	if err := rows.Scan(&base10Id, &createdAt); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	//db.C.Debugf("lastBase36Id: %v", lastBase36Id)
	//db.C.Debugf("base10Id: %v", base10Id)
	//db.C.Debugf("createdAt: %v", createdAt)
	base36Id, err := user.IncrementBase31Id(db, lastBase36Id)
	if err != nil {
		return fmt.Errorf("Error incrementing id: %v", err)
	}
	db.C.Debugf("base36Id: %v", base36Id)
	lastBase36Id = base36Id

	hashmap := make(map[string]interface{})
	hashmap["base10id"] = base10Id
	hashmap["createdAt"] = createdAt
	hashmap["base36id"] = base36Id
	statement := db.NewInsertStatement("base10tobase36", hashmap, "base36id")
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update base10tobase36 table: %v", err)
	}

	return nil
}

func replaceBatch(db *database.DB, newTableName string, newColumnName string) error {
	tableName = newTableName
	columnName = newColumnName

	rows := db.NewStatement(`
		SELECT ` + columnName + `
		FROM ` + tableName + `
		WHERE 1
		`).Query()
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
	row := db.NewStatement(`
		SELECT base36id
		FROM base10tobase36
		WHERE base10id=?`).QueryRow(base10id)
	_, err := row.Scan(&base36id)
	if err != nil {
		return fmt.Errorf("Error reading base36id: %v", err)
	}

	//db.C.Debugf("base36id: %v", base36id)

	hashmap := make(map[string]interface{})
	hashmap[columnName] = base36id

	queryString := `UPDATE ` + tableName + ` SET ` + columnName + ` = "` + base36id + `" WHERE ` + columnName + ` = "` + base10id + `"`
	//db.C.Debugf("queryString: %v", queryString)
	statement := db.NewStatement(queryString)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update table "+tableName+", column "+columnName+", base10id "+base10id+": %v", err)
	}

	return nil
}
