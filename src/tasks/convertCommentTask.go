// convertCommentTask.go converts all the comments into pages.
package tasks

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// ConvertCommentTask is the object that's put into the daemon queue.
type ConvertCommentTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *ConvertCommentTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *ConvertCommentTask) Execute(c sessions.Context) (delay int, err error) {
	delay = tickPeriod
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== COMMENT CONVERSION START ====")
	defer c.Debugf("==== COMMENT CONVERSION COMPLETED SUCCESSFULLY ====")

	// Compute all priors.
	err = database.QuerySql(c, `
		SELECT c.id,c.pageId,c.replyToId,c.createdAt,c.creatorId,c.text,p.groupName
		FROM comments AS c
		JOIN (
			SELECT pageId,groupName
			FROM pages
			WHERE isCurrentEdit
		) AS p
		ON (c.pageId=p.pageId)
		WHERE NOT isConverted`, processComment)
	if err != nil {
		c.Debugf("ERROR: %v", err)
	}
	return 60, err
}

func processComment(c sessions.Context, rows *sql.Rows) error {
	var id, pageId, replyToId, creatorId int64
	var createdAt, text, groupName string
	if err := rows.Scan(&id, &pageId, &replyToId, &createdAt, &creatorId, &text, &groupName); err != nil {
		return fmt.Errorf("failed to scan for commentId: %v", err)
	}
	c.Debugf("==TICK: processing comment: %d", id)

	// Fix line breaks
	text = strings.Replace(text, "\n", "  \n", -1)

	// Try to extract the summary out of the text.
	re := regexp.MustCompile("(?ms)^ {0,3}Summary ?: *\n?(.+?)(\n$|\\z)")
	submatches := re.FindStringSubmatch(text)
	summary := ""
	if len(submatches) > 0 {
		summary = strings.TrimSpace(submatches[1])
	} else {
		// If no summary tags, just extract the first line.
		re := regexp.MustCompile("^(.*)")
		submatches := re.FindStringSubmatch(text)
		summary = strings.TrimSpace(submatches[1])
	}
	// Create parent string
	parentsStr := ""
	if replyToId <= 0 {
		parentsStr = strconv.FormatInt(pageId, 36)
	} else {
		parentsStr = fmt.Sprintf("%s,%s", strconv.FormatInt(pageId, 36), strconv.FormatInt(replyToId, 36))
	}

	// Begin the transaction.
	tx, err := database.NewTransaction(c)
	if err != nil {
		return err
	}

	// Create a new edit.
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = id
	hashmap["creatorId"] = creatorId
	hashmap["title"] = ""
	hashmap["text"] = text
	hashmap["summary"] = summary
	hashmap["alias"] = fmt.Sprintf("%d", id)
	hashmap["sortChildrenBy"] = "chronological"
	hashmap["edit"] = 0
	hashmap["isCurrentEdit"] = 1
	hashmap["type"] = "comment"
	hashmap["groupName"] = groupName
	hashmap["parents"] = parentsStr
	hashmap["createdAt"] = createdAt
	query := database.GetInsertSql("pages", hashmap)
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return err
	}

	// Mark comments as converted
	query = fmt.Sprintf(`UPDATE comments SET isConverted=1 WHERE id=%d`, id)
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return err
	}

	// Add page pair
	hashmap = make(map[string]interface{})
	hashmap["parentId"] = pageId
	hashmap["childId"] = id
	query = database.GetInsertSql("pagePairs", hashmap, "parentId", "childId")
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return err
	}

	if replyToId > 0 {
		hashmap = make(map[string]interface{})
		hashmap["parentId"] = replyToId
		hashmap["childId"] = id
		query = database.GetInsertSql("pagePairs", hashmap, "parentId", "childId")
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
