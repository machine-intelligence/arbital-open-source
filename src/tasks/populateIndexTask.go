// populateIndexTask.go adds all the pages to the index.
package tasks

import (
	"database/sql"
	"fmt"

	"appengine/search"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// PageIndexDoc describes the document which goes into the pages search index.
type PageIndexDoc struct {
	PageId    search.Atom `json:"pageId"`
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Text      string      `json:"text"`
	Alias     search.Atom `json:"alias"`
	GroupName search.Atom `json:"groupName"`
}

// PopulateIndexTask is the object that's put into the daemon queue.
type PopulateIndexTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *PopulateIndexTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *PopulateIndexTask) Execute(c sessions.Context) (delay int, err error) {
	delay = tickPeriod
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== POPULATE INDEX START ====")
	defer c.Debugf("==== POPULATE INDEX COMPLETED SUCCESSFULLY ====")

	// Compute all priors.
	err = database.QuerySql(c, `
		SELECT pageId,type,title,text,alias,groupName
		FROM pages
		WHERE isCurrentEdit`, populateIndexProcessPage)
	if err != nil {
		c.Debugf("ERROR: %v", err)
	}
	return 0, err
}

func populateIndexProcessPage(c sessions.Context, rows *sql.Rows) error {
	var pageId int64
	var alias, groupName string
	p := &PageIndexDoc{}
	if err := rows.Scan(&pageId, &p.Type, &p.Title, &p.Text, &alias, &groupName); err != nil {
		return fmt.Errorf("failed to scan for page: %v", err)
	}
	// Set doc's properties. Our SQL driver doesn't allow to scan strings directly into search.Atom fields
	p.Alias = search.Atom(alias)
	p.GroupName = search.Atom(groupName)
	p.PageId = search.Atom(fmt.Sprintf("%d", pageId))

	index, err := search.Open("pages")
	if err != nil {
		return fmt.Errorf("failed to open index: %v", err)
	}
	_, err = index.Put(c, string(p.PageId), p)
	if err != nil {
		return fmt.Errorf("failed to put page into index: %v", err)
	}

	return nil
}
