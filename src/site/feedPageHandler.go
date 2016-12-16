// feedPageHandler.go serves the feed data for current user.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	FeedPageID = "6rl"
)

type feedData struct {
}

var feedPageHandler = siteHandler{
	URI:         "/json/feed/",
	HandlerFunc: feedPageHandlerFunc,
	Options:     pages.PageOptions{},
}

func feedPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data feedData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load feed rows
	feedRows := make([]*core.FeedSubmission, 0)
	rows := database.NewQuery(`
		SELECT domainId,pageId,submitterId,createdAt
		FROM feedPages
		ORDER BY createdAt DESC
		LIMIT 25`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row core.FeedSubmission
		err := rows.Scan(&row.DomainID, &row.PageID, &row.SubmitterID, &row.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan a FeedSubmission: %v", err)
		}
		core.AddPageToMap(row.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		feedRows = append(feedRows, &row)
		return nil
	})

	// Load claim rows
	claimRows := make([]*core.FeedSubmission, 0)
	rows = database.NewQuery(`
		SELECT editDomainId,pageId,createdBy,createdAt
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		WHERE hasVote
		ORDER BY createdAt DESC
		LIMIT 25`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row core.FeedSubmission
		err := rows.Scan(&row.DomainID, &row.PageID, &row.SubmitterID, &row.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan a FeedSubmission: %v", err)
		}
		core.AddPageToMap(row.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		claimRows = append(claimRows, &row)
		return nil
	})

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["feedRows"] = feedRows
	returnData.ResultMap["claimRows"] = claimRows
	return pages.Success(returnData)
}
