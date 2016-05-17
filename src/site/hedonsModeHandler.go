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

type NewLikesRow struct {
	Names     []string `json:"names"`
	PageId    string   `json:"pageId"`
	ForEdit   bool     `json:"forEdit"`
	CreatedAt string   `json:"createdAt"`
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
	returnData.ResultMap["newLikes"], err = loadNewLikes(db, u, returnData.PageMap)
	if err != nil {
		return pages.Fail("Error loading new likes", err)
	}

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

func loadNewLikes(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page) ([]*NewLikesRow, error) {
	newLikesRows := make([]*NewLikesRow, 0)
	newLikesMap := make(map[string]*NewLikesRow, 0)

	rows := database.NewQuery(`
		SELECT u.Id,u.firstName,u.lastName,pi.pageId,pi.type,l.createdAt,l.value
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN likes AS l
	    ON pi.likeableId=l.likeableId
	    JOIN users AS u
	    ON l.userId=u.id
		WHERE pi.createdBy=?
		ORDER BY l.createdAt DESC`, u.Id).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likerId string
		var firstName string
		var lastName string
		var pageId string
		var pageType string
		var createdAt string
		var likeValue int

		err := rows.Scan(&likerId, &firstName, &lastName, &pageId, &pageType, &createdAt, &likeValue)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if likeValue != 1 || likerId == u.Id {
			return nil
		}

		var newLikesRow *NewLikesRow
		if _, ok := newLikesMap[pageId]; ok {
			newLikesRow = newLikesMap[pageId]
		} else {
			newLikesRow = &NewLikesRow{
				PageId:    pageId,
				CreatedAt: createdAt,
			}
			newLikesMap[pageId] = newLikesRow
			newLikesRows = append(newLikesRows, newLikesRow)

			loadOptions := core.TitlePlusLoadOptions
			if pageType == core.CommentPageType {
				loadOptions.Add(&core.PageLoadOptions{Parents: true})
			}
			core.AddPageToMap(pageId, pageMap, loadOptions)
		}

		newLikesRow.Names = append(newLikesRow.Names, firstName+" "+lastName)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newLikesRows, nil
}
