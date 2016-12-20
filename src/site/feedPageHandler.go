// feedPageHandler.go serves the feed data for current user.

package site

import (
	"encoding/json"
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
	feedRows := make([]*core.FeedPage, 0)
	queryPart := database.NewQuery(`
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (fp.pageId = pi.pageId)
		WHERE NOT pi.hasVote
		ORDER BY fp.score DESC
		LIMIT 25`)
	err = core.LoadFeedPages(db, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		feedRows = append(feedRows, feedPage)
		return nil
	})

	// Load claim rows
	claimRows := make([]*core.FeedPage, 0)
	queryPart = database.NewQuery(`
		JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		ON (fp.pageId = pi.pageId)
		WHERE pi.hasVote
		ORDER BY fp.score DESC
		LIMIT 25`)
	err = core.LoadFeedPages(db, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		claimRows = append(claimRows, feedPage)
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
