// populateElasticTask.go adds all the pages to the elastic index.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
)

// PopulateElasticTask is the object that's put into the daemon queue.
type PopulateElasticTask struct {
}

func (task PopulateElasticTask) Tag() string {
	return "populateElastic"
}

// Check if this task is valid, and we can safely execute it.
func (task PopulateElasticTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task PopulateElasticTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	delay = tickPeriod
	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== POPULATE ELASTIC START ====")
	defer c.Infof("==== POPULATE ELASTIC COMPLETED ====")

	// Delete the index
	err = elastic.DeletePageIndex(c)
	if err != nil {
		// This could happen if we didn't have an index to start with, so we'll go on.
		c.Infof("Couldn't delete page index: %v", err)
	}

	// Create the index
	err = elastic.CreatePageIndex(c)
	if err != nil {
		c.Infof("Couldn't create page index: %v", err)
		return 0, err
	}

	// Compute all priors.
	rows := database.NewQuery(`
		SELECT p.pageId,pi.type,p.title,p.clickbait,p.text,pi.alias,pi.seeDomainId,p.creatorId,pi.externalUrl
		FROM pages AS p
		JOIN pageInfos AS pi
		ON (p.pageId=pi.pageId)
		WHERE p.isLiveEdit
			AND`).AddPart(core.PageInfosFilter(nil)).ToStatement(db).Query()
	err = rows.Process(populateElasticProcessPage)
	if err != nil {
		c.Errorf("ERROR: %v", err)
		// Error or not, we don't want to rerun this.
	}
	return 0, err
}

func populateElasticProcessPage(db *database.DB, rows *database.Rows) error {
	doc := &elastic.Document{}
	if err := rows.Scan(&doc.PageID, &doc.Type, &doc.Title, &doc.Clickbait,
		&doc.Text, &doc.Alias, &doc.SeeDomainID, &doc.CreatorID, &doc.ExternalUrl); err != nil {
		return fmt.Errorf("failed to scan for page: %v", err)
	}

	// Load search strings
	/*pageMap := make(map[string]*core.Page)
	p := core.AddPageIDToMap(doc.PageID, pageMap)
	if err := core.LoadSearchStrings(db, pageMap); err != nil {
		return fmt.Errorf("LoadSearchStrings failed: %v", err)
	}
	doc.SearchStrings = make([]string, 0)
	for _, str := range p.SearchStrings {
		doc.SearchStrings = append(doc.SearchStrings, str)
	}*/

	return elastic.AddPageToIndex(db.C, doc)
}
