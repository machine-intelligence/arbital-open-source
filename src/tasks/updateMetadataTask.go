// updateMetadataTask.go updates all pages to update certain metadata.
package tasks

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

var (
	baseTmpls = []string{"tmpl/scripts.tmpl", "tmpl/style.tmpl"}
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
func (task *UpdateMetadataTask) Execute(c sessions.Context) (delay int, err error) {
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== UPDATE METADATA START ====")
	defer c.Debugf("==== UPDATE METADATA COMPLETED ====")

	// Compute all priors.
	err = database.QuerySql(c, `
		SELECT pageId,edit,text
		FROM pages
		WHERE isCurrentEdit`, updateMetadata)
	if err != nil {
		c.Debugf("ERROR: %v", err)
		return -1, err
	}
	return 0, err
}

func updateMetadata(c sessions.Context, rows *sql.Rows) error {
	var pageId, edit int64
	var text string
	if err := rows.Scan(&pageId, &edit, &text); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	// Begin the transaction.
	tx, err := database.NewTransaction(c)
	if err != nil {
		return err
	}

	// Update page summary
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["edit"] = edit
	hashmap["summary"] = core.ExtractSummary(text)
	hashmap["todoCount"] = core.ExtractTodoCount(text)
	query := database.GetInsertSql("pages", hashmap, "summary", "todoCount")
	if _, err := tx.Exec(query); err != nil {
		tx.Rollback()
		return fmt.Errorf("Couldn't update pages table: %v", err)
	}

	// Update page links table
	err = core.UpdatePageLinks(c, tx, pageId, text, sessions.GetDomain())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Couldn't update links: %v", err)
	}

	// Commit transaction.
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
