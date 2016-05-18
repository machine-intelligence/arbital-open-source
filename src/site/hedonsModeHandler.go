// Handles queries for hedons updates (like 'Alexei liked your page').
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type hedonsModeData struct{}

const (
	// Possible types for hedons rows
	LikesRowType   = "likes"
	ReqsTaughtType = "reqsTaught"
)

type HedonsRow struct {
	Type            string          `json:"type"`
	NewActivityAt   string          `json:"newActivityAt"`
	UserIdsMap      map[string]bool `json:"userIdsMap"` // Unique userIds
	PageId          string          `json:"pageId"`
	RequisiteIdsMap map[string]bool `json:"requisiteIdsMap"` // Unique requisiteIds. Only present if type == ReqsTaughtType
}

var hedonsModeHandler = siteHandler{
	URI:         "/json/hedons/",
	HandlerFunc: hedonsModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func hedonsModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data hedonsModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load new likes on my pages and comments
	likesRows, err := loadReceivedLikes(db, u, returnData.PageMap, returnData.UserMap)
	if err != nil {
		return pages.Fail("Error loading new likes", err)
	}

	// Load requisites taught
	reqsTaughtRows, err := loadRequisitesTaught(db, u, returnData.PageMap, returnData.UserMap)
	if err != nil {
		return pages.Fail("Error loading requisites taught", err)
	}

	returnData.ResultMap["hedons"] = append(likesRows, reqsTaughtRows...)

	// Load and update lastAchievementsView for this user
	returnData.ResultMap[LastAchievementsModeView], err = LoadAndUpdateLastView(db, u, LastAchievementsModeView)
	if err != nil {
		return pages.Fail("Error updating last achievements view", err)
	}

	// Uncomment this to test the feature.
	// returnData.ResultMap[LastAchievementsView] = "2016-05-03 20:11:42"

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func loadReceivedLikes(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, userMap map[string]*core.User) ([]*HedonsRow, error) {
	hedonsRowMap := make(map[string]*HedonsRow, 0)

	rows := database.NewQuery(`
		SELECT u.Id,pi.pageId,pi.type,l.updatedAt,l.value
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN likes AS l
		ON pi.likeableId=l.likeableId
		JOIN users AS u
		ON l.userId=u.id
		WHERE pi.createdBy=?`, u.Id).Add(` AND l.value=1 AND l.userId!=?`, u.Id).Add(`
		ORDER BY l.updatedAt DESC`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likerId string
		var pageId string
		var pageType string
		var updatedAt string
		var likeValue int

		err := rows.Scan(&likerId, &pageId, &pageType, &updatedAt, &likeValue)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		hedonsRow, ok := hedonsRowMap[pageId]
		if !ok {
			hedonsRow = &HedonsRow{
				PageId:        pageId,
				NewActivityAt: updatedAt,
				Type:          LikesRowType,
				UserIdsMap:    make(map[string]bool, 0),
			}
			hedonsRowMap[pageId] = hedonsRow

			loadOptions := core.TitlePlusLoadOptions
			if pageType == core.CommentPageType {
				loadOptions.Add(&core.PageLoadOptions{Parents: true})
			}
			core.AddPageToMap(pageId, pageMap, loadOptions)
		}

		core.AddUserIdToMap(likerId, userMap)
		hedonsRow.UserIdsMap[likerId] = true

		return nil
	})
	if err != nil {
		return nil, err
	}

	hedonsRows := make([]*HedonsRow, 0, len(hedonsRowMap))
	for _, row := range hedonsRowMap {
		hedonsRows = append(hedonsRows, row)
	}

	return hedonsRows, nil
}

// Load all the requisites taught by this user.
func loadRequisitesTaught(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, userMap map[string]*core.User) ([]*HedonsRow, error) {
	hedonsRowMap := make(map[string]*HedonsRow, 0)

	rows := database.NewQuery(`
		SELECT u.Id,pi.pageId,ump.masteryId,ump.updatedAt
		FROM userMasteryPairs AS ump
		JOIN `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		ON ump.taughtBy=pi.pageId
		JOIN users AS u
		ON ump.userId=u.id
		WHERE pi.createdBy=?`, u.Id).Add(` AND ump.has=1 AND ump.userId!=?`, u.Id).Add(`
		ORDER BY ump.updatedAt DESC`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var learnerId string
		var learnerName string
		var taughtById string
		var masteryId string
		var updatedAt string

		err := rows.Scan(&learnerId, &taughtById, &masteryId, &updatedAt)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if learnerId == u.Id {
			return nil
		}

		hedonsRow, ok := hedonsRowMap[taughtById]
		if !ok {
			hedonsRow = &HedonsRow{
				PageId:          taughtById,
				NewActivityAt:   updatedAt,
				Type:            ReqsTaughtType,
				UserIdsMap:      make(map[string]bool, 0),
				RequisiteIdsMap: make(map[string]bool, 0),
			}
			hedonsRowMap[taughtById] = hedonsRow
		}

		core.AddPageToMap(taughtById, pageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap(masteryId, pageMap, core.TitlePlusLoadOptions)
		core.AddUserIdToMap(learnerId, userMap)
		hedonsRow.UserIdsMap[learnerName] = true
		hedonsRow.RequisiteIdsMap[masteryId] = true

		return nil
	})
	if err != nil {
		return nil, err
	}

	hedonsRows := make([]*HedonsRow, 0, len(hedonsRowMap))
	for _, row := range hedonsRowMap {
		hedonsRows = append(hedonsRows, row)
	}

	return hedonsRows, nil
}
