// feedHandler.go serves the feed data for current user.

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

type FeedRow struct {
	DomainID    string `json:"domainId"`
	PageID      string `json:"pageId"`
	SubmitterID string `json:"submitterId"`
	CreatedAt   string `json:"createdAt"`
}

type feedData struct {
}

var feedHandler = siteHandler{
	URI:         "/json/feed/",
	HandlerFunc: feedHandlerFunc,
	Options:     pages.PageOptions{},
}

func feedHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data feedData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	feedRows := make([]*FeedRow, 0)
	rows := database.NewQuery(`
		SELECT domainId,pageId,submitterId,createdAt
		FROM feedPages
		ORDER BY createdAt DESC
		LIMIT 25`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row FeedRow
		err := rows.Scan(&row.DomainID, &row.PageID, &row.SubmitterID, &row.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan a FeedRow: %v", err)
		}
		core.AddPageToMap(row.PageID, returnData.PageMap, core.TitlePlusLoadOptions)
		feedRows = append(feedRows, &row)
		return nil
	})

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["feedRows"] = feedRows
	return pages.Success(returnData)
}
