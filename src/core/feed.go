// pageUtils.go contains various helpers for dealing with pages
package core

import (
	"fmt"
	"time"

	"zanaduu3/src/database"
)

const (
	// Feed page score values
	LikeFeedPageScore    = 10
	CommentFeedPageScore = 50 * LikeFeedPageScore
	VoteFeedPageScore    = 10 * LikeFeedPageScore
	NewFeedPageScore     = 30 * CommentFeedPageScore
)

// A page that was submitted to a domain feed
type FeedPage struct {
	DomainID    string  `json:"domainId"`
	PageID      string  `json:"pageId"`
	SubmitterID string  `json:"submitterId"`
	CreatedAt   string  `json:"createdAt"`
	Score       float64 `json:"score"`
}

// Compute the factor to multiple with a score, such that the score decays based on
// the duration since the given date
func _getScoreFactor(createdAt string) float64 {
	t, _ := time.Parse(database.TimeLayout, createdAt)
	days := time.Since(t).Hours() / 24
	if days < 1 {
		return 1
	}
	return 1 / days
}

// Compute the score of the given feed page
func (fp *FeedPage) ComputeScore(pageMap map[string]*Page) {
	p := pageMap[fp.PageID]
	fp.Score = NewFeedPageScore * _getScoreFactor(fp.CreatedAt)
	fp.Score += LikeFeedPageScore * float64(p.LikeCount)
	for _, commentID := range p.CommentIDs {
		fp.Score += CommentFeedPageScore * _getScoreFactor(pageMap[commentID].PageCreatedAt)
	}
	for _, vote := range p.Votes {
		fp.Score += VoteFeedPageScore * _getScoreFactor(vote.CreatedAt)
	}
}

type ProcessLoadFeedPageCallback func(db *database.DB, feedPage *FeedPage) error

// LoadFeedPages loads rows from feedPages table
func LoadFeedPages(db *database.DB, queryPart *database.QueryPart, callback ProcessLoadFeedPageCallback) error {
	rows := database.NewQuery(`
		SELECT fp.domainId, fp.pageId, fp.submitterId, fp.createdAt, fp.score
		FROM feedPages AS fp`).AddPart(queryPart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var row FeedPage
		err := rows.Scan(&row.DomainID, &row.PageID, &row.SubmitterID, &row.CreatedAt, &row.Score)
		if err != nil {
			return fmt.Errorf("Failed to scan a feed page: %v", err)
		}
		return callback(db, &row)
	})
	return err
}

// LoadFeedSubmissionsForPages loads all the times the given pages have been submitted to a feed
func LoadFeedSubmissionsForPages(db *database.DB, u *CurrentUser, resultData *CommonHandlerData, options *LoadDataOptions) error {
	sourcePageMap := options.ForPages
	pageIDs := PageIDsListFromMap(sourcePageMap)
	if len(pageIDs) <= 0 {
		return nil
	}

	queryPart := database.NewQuery(`WHERE fp.pageId IN`).AddArgsGroup(pageIDs)
	err := LoadFeedPages(db, queryPart, func(db *database.DB, feedPage *FeedPage) error {
		AddUserIDToMap(feedPage.SubmitterID, resultData.UserMap)
		sourcePageMap[feedPage.PageID].FeedSubmissions = append(sourcePageMap[feedPage.PageID].FeedSubmissions, feedPage)
		if _, ok := resultData.DomainMap[feedPage.DomainID]; !ok {
			resultData.DomainMap[feedPage.DomainID] = NewDomainWithID(feedPage.DomainID)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Couldn't load feed submissions for pages: %v", err)
	}
	return nil
}
