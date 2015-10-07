// fixTextTask.go updates all pages' text fields to fix common mistakes
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/database"
)

// FixTextTask is the object that's put into the daemon queue.
type FixTextTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *FixTextTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *FixTextTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== FIX TEXT START ====")
	defer c.Debugf("==== FIX TEXT COMPLETED ====")

	rows := db.NewStatement(`
		SELECT pageId,edit,text
		FROM pages
		WHERE isCurrentEdit`).Query()
	if err = rows.Process(fixText); err != nil {
		c.Debugf("ERROR, failed to fix text: %v", err)
		return 0, err
	}
	return 0, err
}

func fixText(db *database.DB, rows *database.Rows) error {
	var pageId, edit int64
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

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
	return nil
}
