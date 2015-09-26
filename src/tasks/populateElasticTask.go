// populateElasticTask.go adds all the pages to the elastic index.
package tasks

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/sessions"
)

// PopulateElasticTask is the object that's put into the daemon queue.
type PopulateElasticTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *PopulateElasticTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *PopulateElasticTask) Execute(c sessions.Context) (delay int, err error) {
	delay = tickPeriod
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== POPULATE ELASTIC START ====")
	defer c.Debugf("==== POPULATE ELASTIC COMPLETED ====")

	// Delete the index
	err = elastic.DeletePageIndex(c)
	if err != nil {
		// This could happen if we didn't have an index to start with, so we'll go on.
		c.Debugf("Couldn't delete page index: %v", err)
	}

	// Create the index
	err = elastic.CreatePageIndex(c)
	if err != nil {
		c.Debugf("Couldn't create page index: %v", err)
		return 0, err
	}

	// Compute all priors.
	err = database.QuerySql(c, `
		SELECT pageId,type,title,clickbait,text,alias,groupId,creatorId
		FROM pages
		WHERE isCurrentEdit`, populateElasticProcessPage)
	if err != nil {
		c.Debugf("ERROR: %v", err)
		// Error or not, we don't want to rerun this.
	}
	return 0, err
}

func populateElasticProcessPage(c sessions.Context, rows *sql.Rows) error {
	doc := &elastic.Document{}
	if err := rows.Scan(&doc.PageId, &doc.Type, &doc.Title, &doc.Clickbait,
		&doc.Text, &doc.Alias, &doc.GroupId, &doc.CreatorId); err != nil {
		return fmt.Errorf("failed to scan for page: %v", err)
	}

	return elastic.AddPageToIndex(c, doc)
}
