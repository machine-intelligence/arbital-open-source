// discussionModeHandler.go serves the /discussion panel.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type discussionModeData struct {
	NumPagesToLoad int
}

var discussionModeHandler = siteHandler{
	URI:         "/json/discussionMode/",
	HandlerFunc: discussionModeHandlerFunc,
	Options:     pages.PageOptions{},
}

func discussionModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data discussionModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad == 0 {
		data.NumPagesToLoad = 25
	}

	// Load all comments of interest
	returnData.ResultMap["commentIds"], err = loadDiscussions(db, u, returnData.PageMap, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("failed to load hot page ids", err)
	}

	// Load and update LastDiscussionView for this user
	returnData.ResultMap[LastDiscussionModeView], err = LoadAndUpdateLastView(db, u, LastDiscussionModeView)
	if err != nil {
		return pages.Fail("Error updating last read mode view", err)
	}

	// Load the pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func loadDiscussions(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, numPagesToLoad int) ([]string, error) {
	commentIds := make([]string, 0)
	parentPageOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)
	childPageOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.TitlePlusLoadOptions)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		JOIN pagePairs AS pp
		ON (pp.childId=pi.pageId)
		JOIN subscriptions AS s
		ON (pp.parentId=s.toId)
		WHERE s.userId=?`, u.Id).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
		GROUP BY pp.childId
		ORDER BY pi.createdAt DESC
		LIMIT ?`, numPagesToLoad).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		commentIds = append(commentIds, childId)
		core.AddPageToMap(parentId, pageMap, parentPageOptions)
		core.AddPageToMap(childId, pageMap, childPageOptions)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error reading rows: %v", err)
	}

	return commentIds, nil
}
