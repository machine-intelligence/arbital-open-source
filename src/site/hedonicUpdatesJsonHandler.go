// Handles queries for hedonic updates (like 'Alexei liked your page').
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type hedonicUpdatesJsonData struct{}

type NewLikesRow struct {
	Names            []string `json:"names"`
	PageId           string   `json:"pageId"`
	ForEdit          bool     `json:"forEdit"`
	CreatedAt        string   `json:"createdAt"`
	NewSinceLastView bool     `json:"newSinceLastView"`
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
	returnData := core.NewHandlerData(u)

	// Decode data
	var data hedonicUpdatesJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Load lastAchievementsView for this user
	var lastView string
	row := database.NewQuery(`
		SELECT lastAchievementsView
		FROM lastViews
		WHERE userId=?`, u.Id).ToStatement(db).QueryRow()
	_, err = row.Scan(&lastView)
	if err != nil {
		return pages.Fail("Couldn't load lastAchievementsView", err).Status(http.StatusBadRequest)
	}

	// Update lastAchievementsView for this user
	hashmap := make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["lastAchievementsView"] = database.Now()
	statement := db.NewInsertStatement("lastViews", hashmap, "lastAchievementsView")
	if _, err := statement.Exec(); err != nil {
		return pages.Fail("Couldn't update lastAchievementsView", err).Status(http.StatusBadRequest)
	}

	// Load new likes on my pages and comments
	returnData.ResultMap["newLikes"], err = loadNewLikes(db, u, returnData.PageMap, lastView)
	if err != nil {
		return pages.Fail("Error loading new likes", err).Status(http.StatusBadRequest)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func loadNewLikes(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, lastView string) ([]*NewLikesRow, error) {
	newLikesRows := make([]*NewLikesRow, 0)
	newLikesMap := make(map[string]*NewLikesRow, 0)

	rows := database.NewQuery(`
		SELECT u.Id,u.firstName,u.lastName,pi.pageId,l.createdAt,l.value
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
		var createdAt string
		var likeValue int

		err := rows.Scan(&likerId, &firstName, &lastName, &pageId, &createdAt, &likeValue)
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
				PageId:           pageId,
				CreatedAt:        createdAt,
				NewSinceLastView: createdAt > lastView,
			}
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
