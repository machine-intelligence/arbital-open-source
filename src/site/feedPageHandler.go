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
	MinFeaturedCommentTextLength = 20
)

type feedData struct {
	FilterByTagAlias string `json:"filterByTagAlias"`
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

	filterByTagID := ""
	if len(data.FilterByTagAlias) > 0 {
		// Get actual tag id
		tagID, ok, err := core.LoadAliasToPageID(db, u, data.FilterByTagAlias)
		if err != nil {
			return pages.Fail("Couldn't convert alias", err)
		}
		if !ok {
			return pages.Fail("Couldn't find tag", err)
		}
		filterByTagID = tagID
	}

	feedRowLoadOptions := (&core.PageLoadOptions{
		Tags: true,
	}).Add(core.IntrasitePopoverLoadOptions)

	tagFilter := database.NewQuery("")
	if len(filterByTagID) > 0 {
		tagFilter = database.NewQuery(`
		JOIN pagePairs AS pp
		ON pi.pageId = pp.childId
			AND pp.type = ?`, core.TagPagePairType).Add(`
			AND pp.parentId = ?`, filterByTagID)
	}
	orderByAndLimit := database.NewQuery(`
		ORDER BY fp.score DESC
		LIMIT 25`)

	// Load feed rows
	feedRows := make([]*core.FeedPage, 0)
	queryPart := database.NewQuery(`AND NOT pi.hasVote`).AddPart(tagFilter).AddPart(orderByAndLimit)

	err = core.LoadFeedPages(db, u, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, feedRowLoadOptions)
		feedRows = append(feedRows, feedPage)
		return nil
	})
	if err != nil {
		return pages.Fail("Couldn't load feed rows", err)
	}

	// Load claim rows
	claimRows := make([]*core.FeedPage, 0)
	queryPart = database.NewQuery(`AND pi.hasVote`).AddPart(tagFilter).AddPart(orderByAndLimit)

	err = core.LoadFeedPages(db, u, queryPart, func(db *database.DB, feedPage *core.FeedPage) error {
		core.AddPageToMap(feedPage.PageID, returnData.PageMap, feedRowLoadOptions)
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

	if len(claimIDs) > 0 {
		// For each claim row, find the best comment to show
		rows := database.NewQuery(`
			SELECT t.parentId,t.childId
			FROM (
				SELECT pp.parentId,pp.childId
				FROM pagePairs AS pp
				JOIN pageInfos AS pi
				ON pp.parentId IN`).AddArgsGroupStr(claimIDs).Add(`
					AND pp.childId = pi.pageId
					AND pi.type=?`, core.CommentPageType).Add(`
					AND NOT pi.isResolved AND NOT pi.isEditorComment AND pi.isApprovedComment
				JOIN pages AS p
				ON pi.pageId = p.pageId AND p.isLiveEdit
					AND length(p.text) >= ?`, MinFeaturedCommentTextLength).Add(`
					/* No replies */
					/*AND pp.childId IN (SELECT childId FROM pagePairs GROUP BY 1 HAVING SUM(1) <= 1)*/
					AND`).AddPart(core.PageInfosFilter(u)).Add(`
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
