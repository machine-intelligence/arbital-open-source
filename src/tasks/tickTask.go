// tick.go updates all the computed values in our database.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	tickPeriod = 5 * 60 // 5 minutes
)

type FeedRowScore struct {
}

// TickTask is the object that's put into the daemon queue.
type TickTask struct {
}

func (task TickTask) Tag() string {
	return "tick"
}

// Check if this task is valid, and we can safely execute it.
func (task TickTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task TickTask) Execute(db *database.DB) (delay int, err error) {
	delay = tickPeriod
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	// Update pages' view counts
	/*query := database.NewQuery(`
		UPDATE pageInfos AS pi
		SET pi.viewCount=(
			SELECT COUNT(DISTINCT userId)
			FROM visits AS v
			WHERE v.pageId=pi.pageId
		)`).ToStatement(db)
	if _, err := query.Exec(); err != nil {
		c.Errorf("Failed to update view count: %v", err)
	}*/

	feedPages := make([]*core.FeedPage, 0)
	returnData := core.NewHandlerData(core.NewCurrentUser())

	// Load all feeed pages
	queryPart := database.NewQuery(``)
	err = core.LoadFeedPages(db, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, (&core.PageLoadOptions{
			Votes:    true,
			Comments: true,
		}).Add(core.TitlePlusLoadOptions))
		feedPages = append(feedPages, feedPage)
		return nil
	})
	if err != nil {
		err = fmt.Errorf("Failed to update view count: %v", err)
		return
	}

	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		err = fmt.Errorf("Pipeline error: %v", err)
		return
	}

	hashmaps := make(database.InsertMaps, 0)
	for _, feedPage := range feedPages {
		feedPage.ComputeScore(returnData.PageMap)
		hashmap := make(database.InsertMap)
		hashmap["domainId"] = feedPage.DomainID
		hashmap["pageId"] = feedPage.PageID
		hashmap["score"] = feedPage.Score
		hashmaps = append(hashmaps, hashmap)
	}
	statement := db.NewMultipleInsertStatement("feedPages", hashmaps, "score")
	if _, err = statement.Exec(); err != nil {
		err = fmt.Errorf("Couldn't insert into lastVisits: %v", err)
		return
	}

	c.Infof("==== TICK START ====")
	defer c.Infof("==== TICK COMPLETED ====")
	return
}
