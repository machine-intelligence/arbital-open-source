// Handles queries for hedonic updates (like 'Alexei liked your page').
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type hedonicUpdatesJsonData struct{}

type NewLikesRow struct {
	Names     []string `json:"names"`
	PageId    string   `json:"pageId"`
	ForEdit   bool     `json:"forEdit"`
	CreatedAt string   `json:"createdAt"`
}

var hedonicUpdatesJsonHandler = siteHandler{
	URI:         "/json/hedons/",
	HandlerFunc: hedonicUpdatesHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func hedonicUpdatesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data hedonicUpdatesJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Load new likes on my pages, edits, and comments
	returnData.ResultMap["newLikes"], err = loadNewLikes(db, u, returnData.PageMap)
	if err != nil {
		return pages.HandlerBadRequestFail("Error loading new likes", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData)
}

func loadNewLikes(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page) ([]*NewLikesRow, error) {
	newLikesRows := make([]*NewLikesRow, 0)
	newLikesMap := make(map[string]*NewLikesRow, 0)

	rows := database.NewQuery(`
		SELECT u.firstName,u.lastName,pi.pageId,l.createdAt,l.value
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN likes AS l
	    ON pi.likeableId=l.likeableId
	    JOIN users AS u
	    ON l.userId=u.id
		WHERE pi.createdBy=?
		LIMIT 100`, u.Id).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var firstName string
		var lastName string
		var pageId string
		var createdAt string
		var likeValue int

		err := rows.Scan(&firstName, &lastName, &pageId, &createdAt, &likeValue)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		if likeValue != 1 {
			return nil
		}

		var newLikesRow *NewLikesRow
		if _, ok := newLikesMap[pageId]; ok {
			newLikesRow = newLikesMap[pageId]
		} else {
			newLikesRow = &NewLikesRow{PageId: pageId, CreatedAt: createdAt}
			newLikesMap[pageId] = newLikesRow
			newLikesRows = append(newLikesRows, newLikesRow)
			core.AddPageToMap(pageId, pageMap, core.TitlePlusLoadOptions)
		}

		newLikesRow.Names = append(newLikesRow.Names, firstName+" "+lastName)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newLikesRows, nil
}
