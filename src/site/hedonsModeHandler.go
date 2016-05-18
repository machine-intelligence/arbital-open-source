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
	Type          string            `json:"type"`
	NewActivityAt string            `json:"newActivityAt"`
	Names         map[string]string `json:"names"`
	PageId        string            `json:"pageId"`
	RequisiteIds  map[string]string `json:"requisiteIds"` // Only present if type == ReqsTaughtType
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
	likesRows, err := loadReceivedLikes(db, u, returnData.PageMap)
	if err != nil {
		return pages.Fail("Error loading new likes", err)
	}

	// Load requisites taught
	reqsTaughtRows, err := loadRequisitesTaught(db, u, returnData.PageMap)
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

func loadReceivedLikes(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page) ([]*HedonsRow, error) {
	hedonsRowMap := make(map[string]*HedonsRow, 0)

	rows := database.NewQuery(`
		SELECT u.Id,CONCAT(u.firstName," ",u.lastName),pi.pageId,pi.type,l.updatedAt,l.value
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN likes AS l
	    ON pi.likeableId=l.likeableId
	    JOIN users AS u
	    ON l.userId=u.id
		WHERE pi.createdBy=?
			AND l.value=1
		ORDER BY l.updatedAt DESC`, u.Id).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likerId string
		var likerName string
		var pageId string
		var pageType string
		var updatedAt string
		var likeValue int

		err := rows.Scan(&likerId, &likerName, &pageId, &pageType, &updatedAt, &likeValue)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if likerId == u.Id {
			return nil
		}

		var hedonsRow *HedonsRow
		if _, ok := hedonsRowMap[pageId]; ok {
			hedonsRow = hedonsRowMap[pageId]
		} else {
			hedonsRow = &HedonsRow{
				PageId:        pageId,
				NewActivityAt: updatedAt,
				Type:          LikesRowType,
				Names:         make(map[string]string, 0),
			}
			hedonsRowMap[pageId] = hedonsRow

			loadOptions := core.TitlePlusLoadOptions
			if pageType == core.CommentPageType {
				loadOptions.Add(&core.PageLoadOptions{Parents: true})
			}
			core.AddPageToMap(pageId, pageMap, loadOptions)
		}

		hedonsRow.Names[likerName] = likerName

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
func loadRequisitesTaught(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page) ([]*HedonsRow, error) {
	hedonsRowMap := make(map[string]*HedonsRow, 0)

	rows := database.NewQuery(`
	    SELECT u.Id,CONCAT(u.firstName," ",u.lastName),pi.pageId,ump.masteryId,ump.updatedAt
	    FROM userMasteryPairs AS ump
	    JOIN `).AddPart(core.PageInfosTable(u)).Add(` AS pi
	    ON ump.taughtBy=pi.pageId
	    JOIN users AS u
	    ON ump.userId=u.id
	    WHERE pi.createdBy=?
			AND ump.has=1
		ORDER BY ump.updatedAt DESC`, u.Id).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var learnerId string
		var learnerName string
		var taughtById string
		var masteryId string
		var updatedAt string

		err := rows.Scan(&learnerId, &learnerName, &taughtById, &masteryId, &updatedAt)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if learnerId == u.Id {
			return nil
		}

		var hedonsRow *HedonsRow
		if _, ok := hedonsRowMap[taughtById]; ok {
			hedonsRow = hedonsRowMap[taughtById]
		} else {
			hedonsRow = &HedonsRow{
				PageId:        taughtById,
				NewActivityAt: updatedAt,
				Type:          ReqsTaughtType,
				Names:         make(map[string]string, 0),
				RequisiteIds:  make(map[string]string, 0),
			}
			hedonsRowMap[taughtById] = hedonsRow
		}

		core.AddPageToMap(taughtById, pageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap(masteryId, pageMap, core.TitlePlusLoadOptions)
		hedonsRow.Names[learnerName] = learnerName
		hedonsRow.RequisiteIds[masteryId] = masteryId

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
