// updateMetadataTask.go updates all pages to update certain metadata.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// UpdateMetadataTask is the object that's put into the daemon queue.
type UpdateMetadataTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *UpdateMetadataTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *UpdateMetadataTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== UPDATE METADATA START ====")
	defer c.Debugf("==== UPDATE METADATA COMPLETED ====")

	// Regenerate pages and links tables
	rows := db.NewStatement(`
		SELECT pageId,edit,text
		FROM pages
		WHERE isCurrentEdit`).Query()
	if err = rows.Process(updateMetadata); err != nil {
		c.Debugf("ERROR, failed to update pages and pageLinks: %v", err)
		return 0, err
	}

	// Regenerate pageInfos table
	rows = db.NewStatement(`
		SELECT pageId, MAX(if(isCurrentEdit, edit, 0)), MAX(edit), MIN(createdAt)
		FROM pages
		WHERE 1
		GROUP BY pageId`).Query()
	if err = rows.Process(updatePageInfos); err != nil {
		c.Debugf("ERROR, failed to update pageInfos: %v", err)
		return 0, err
	}
	return 0, err
}

func updateMetadata(db *database.DB, rows *database.Rows) error {
	var pageId, edit int64
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	text, err := core.StandardizeLinks(db, text)
	if err != nil {
		return fmt.Errorf("failed to standardize links: %v", err)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Update pages table
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = edit
		hashmap["text"] = text
		hashmap["todoCount"] = core.ExtractTodoCount(text)
		statement := tx.NewInsertTxStatement("pages", hashmap, "text", "todoCount")
		if _, err := statement.Exec(); err != nil {
			return "Couldn't update pages table", err
		}

		// Delete old page summaries
		statement = database.NewQuery(`
			DELETE FROM pageSummaries WHERE pageId=?`, pageId).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return "Couldn't delete existing page summaries", err
		}

		// Add new page summaries
		_, summaryValues := core.ExtractSummaries(pageId, text)
		statement = tx.NewTxStatement(`
			INSERT INTO pageSummaries (pageId,name,text)
			VALUES ` + database.ArgsPlaceholder(len(summaryValues), 3))
		if _, err := statement.Exec(summaryValues...); err != nil {
			return "Couldn't insert page summaries", err
		}

		// Update page links table
		err := core.UpdatePageLinks(tx, pageId, text, sessions.GetDomain())
		if err != nil {
			return "Couldn't update links", err
		}
		return "", nil
	})
	if errMessage != "" {
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	return nil
}

func updatePageInfos(db *database.DB, rows *database.Rows) error {
	var pageId, currentEdit, maxEdit int64
	var createdAt string
	if err := rows.Scan(&pageId, &currentEdit, &maxEdit, &createdAt); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Update pageInfos summary
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["currentEdit"] = currentEdit
		hashmap["maxEdit"] = maxEdit
		hashmap["createdAt"] = createdAt
		statement := tx.NewInsertTxStatement("pageInfos", hashmap, "currentEdit", "maxEdit", "createdAt")
		if _, err := statement.Exec(); err != nil {
			return "Couldn't update pageInfos table", err
		}
		return "", nil
	})
	if errMessage != "" {
		return fmt.Errorf("%s: %v", errMessage, err)
	}

	return nil
}
