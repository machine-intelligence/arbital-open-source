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
	FeedPageID                   = "6rl"
	MinFeaturedCommentTextLength = 140
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
		WHERE NOT pi.hasVote
		ORDER BY fp.score DESC
		LIMIT 25`)
	err = core.LoadFeedPages(db, u, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		feedRows = append(feedRows, feedPage)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load feed rows", err)
	}

	// Load claim rows
	claimRows := make([]*core.FeedPage, 0)
	queryPart = database.NewQuery(`
		WHERE pi.hasVote
		ORDER BY fp.score DESC
		LIMIT 25`)
	err = core.LoadFeedPages(db, u, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, core.IntrasitePopoverLoadOptions)
		claimRows = append(claimRows, feedPage)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load claim rows", err)
	}

	claimMap := make(map[string]*core.FeedPage) // pageId -> feed page
	claimIDs := make([]string, 0)
	for _, claimRow := range claimRows {
		claimIDs = append(claimIDs, claimRow.PageID)
		claimMap[claimRow.PageID] = claimRow
	}

	// For each claim row, find the best comment to show
	rows := database.NewQuery(`
		SELECT t.parentId,t.childId
		FROM (
			SELECT pp.parentId,pp.childId
			FROM pagePairs AS pp
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (pp.childId = pi.pageId)
			JOIN pages AS p
			ON (pi.pageId = p.pageId AND p.isLiveEdit)
			WHERE pp.parentId IN`).AddArgsGroupStr(claimIDs).Add(`
				AND length(p.text) >= ?`, MinFeaturedCommentTextLength).Add(`
				AND pi.type=?`, core.CommentPageType).Add(`
				AND NOT pi.isResolved AND NOT pi.isEditorComment AND pi.isApprovedComment
			ORDER BY pi.createdAt DESC
		) AS t
		GROUP BY 1,2`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var claimPageID, commentID string
		err := rows.Scan(&claimPageID, &commentID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		claimMap[claimPageID].FeaturedCommentID = commentID
		core.AddPageToMap(commentID, returnData.PageMap, core.SubpageLoadOptions)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load featured claims", err)
	}

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["feedRows"] = feedRows
	returnData.ResultMap["claimRows"] = claimRows
	return pages.Success(returnData)
}
