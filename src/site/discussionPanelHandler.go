// discussionPanelHandler.go serves the /discussion panel.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

type discussionPanelData struct {
	NumPagesToLoad int
}

var discussionPanelHandler = siteHandler{
	URI:         "/json/discussionPanel/",
	HandlerFunc: discussionPanelHandlerFunc,
	Options:     pages.PageOptions{},
}

func discussionPanelHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data discussionPanelData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumPagesToLoad == 0 {
		data.NumPagesToLoad = 25
	}

	// Load all comments which have a parent to which you are subscribed
	//returnData.ResultMap["pageIds"], err = loadDiscussions(db, u, returnData.PageMap, numPagesToLoad)
	//if err != nil {
	//return pages.Fail("failed to load hot page ids", err)
	//}

	// Load and update LastDiscussionView for this user
	returnData.ResultMap[LastDiscussionView], err = LoadAndUpdateLastView(db, u, LastDiscussionView)
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

/*func loadDiscussions(db *database.DB, u *core.CurrentUser, pageMap map[string]*core.Page, numPagesToLoad int) ([]string, error) {
	childrenIds := make([]string, 0)
	pageOptions := (&core.PageLoadOptions{SubpageCounts: true}).Add(core.TitlePlusLoadOptions)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId
		FROM`).AddPart(core.PageInfosTable(u)).Add(`AS pi
		JOIN pagePairs AS pp
		ON (pp.childId=pi.pageId)
		JOIN subscriptions AS s
		ON (pp.parentId=s.toId)
		WHERE s.userId=?`, u.Id).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
		ORDER BY pi.createdAt DESC
		LIMIT ?`, numPagesToLoad).ToStatement(db).Query()

	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId string
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		childrenIds = append(childrenIds, childId)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error reading rows: %v", err)
	}

	return nil
}*/
