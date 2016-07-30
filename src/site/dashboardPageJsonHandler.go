// dashboardPage.go serves the dashboard template.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var dashboardPageHandler = siteHandler{
	URI:         "/json/dashboardPage/",
	HandlerFunc: dashboardPageJSONHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

type dashboardPageJSONData struct {
	NumToLoad int
}

// dashboardPageJsonHandler renders the dashboard page.
func dashboardPageJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data dashboardPageJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumToLoad <= 0 {
		data.NumToLoad = indexPanelLimit
	}

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

	_, err = core.LoadAllDomainIDs(db, returnData.PageMap)
	if err != nil {
		return pages.Fail("Error while loading domain ids", err)
	}

	returnData.ResultMap[RecentlyCreatedCommentIds], err = LoadRecentlyCreatedComment(u, params.PrivateGroupID, data.NumToLoad, db, returnData, pageOptions)
	if err != nil {
		return pages.Fail("error while loading "+RecentlyCreatedCommentIds, err)
	}

	returnData.ResultMap[RecentlyEditedIds], err = LoadRecentlyEdited(u, params.PrivateGroupID, data.NumToLoad, db, returnData, pageOptions)
	if err != nil {
		return pages.Fail("error while loading "+RecentlyEditedIds, err)
	}

	returnData.ResultMap[PagesWithDraftIds], err = LoadPagesWithDraft(u, params.PrivateGroupID, data.NumToLoad, db, returnData, pageOptions)
	if err != nil {
		return pages.Fail("error while loading "+PagesWithDraftIds, err)
	}

	returnData.ResultMap[MostTodosIds], err = LoadMostTodos(u, params.PrivateGroupID, data.NumToLoad, db, returnData, pageOptions)
	if err != nil {
		return pages.Fail("error while loading "+MostTodosIds, err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	err = loadStats(db, returnData.ResultMap, u)
	if err != nil {
		return pages.Fail("error loading stats", err)
	}

	return pages.Success(returnData)
}

func loadStats(db *database.DB, resultMap map[string]interface{}, u *core.CurrentUser) error {

	// Load number of wiki pages and comments created by this user
	rows := database.NewQuery(`
		SELECT pi.type,COUNT(*)
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		WHERE pi.createdBy=?
		GROUP BY pi.type`, u.ID).ToStatement(db).Query()
	resultMap["numWikiPages"] = 0
	resultMap["numComments"] = 0
	resultMap["numQuestions"] = 0
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageType string
		var count int
		err := rows.Scan(&pageType, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		switch pageType {
		case core.WikiPageType:
			resultMap["numWikiPages"] = count
		case core.CommentPageType:
			resultMap["numComments"] = count
		case core.QuestionPageType:
			resultMap["numQuestions"] = count
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Load number of likes on wiki pages and comments created by this user
	rows = database.NewQuery(`
		SELECT pi.type,COUNT(*)
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN likes AS l
	    ON pi.likeableId=l.likeableId
		WHERE pi.createdBy=?
		GROUP BY pi.type`, u.ID).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageType string
		var count int
		err = rows.Scan(&pageType, &count)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		switch pageType {
		case core.WikiPageType:
			resultMap["wikiLikes"] = count
		case core.CommentPageType:
			resultMap["commentLikes"] = count
		case core.QuestionPageType:
			resultMap["commentLikes"] = count
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Load number of users taught by this user
	var numUsersTaught int
	row := database.NewQuery(`
	    SELECT COUNT(DISTINCT ump.userId)
		FROM userMasteryPairs AS ump
	    JOIN `).AddPart(core.PageInfosTable(u)).Add(` AS pi
	    ON ump.taughtBy=pi.pageId
	    WHERE pi.createdBy=?`, u.ID).ToStatement(db).QueryRow()
	_, err = row.Scan(&numUsersTaught)
	if err != nil {
		return err
	}
	resultMap["numUsersTaught"] = numUsersTaught

	// Load number of requisites taught by this user
	var numReqsTaught int
	row = database.NewQuery(`
	    SELECT COUNT(*)
	    FROM userMasteryPairs AS ump
	    JOIN `).AddPart(core.PageInfosTable(u)).Add(` AS pi
	    ON ump.taughtBy=pi.pageId
	    WHERE pi.createdBy=?`, u.ID).ToStatement(db).QueryRow()
	_, err = row.Scan(&numReqsTaught)
	if err != nil {
		return err
	}
	resultMap["numReqsTaught"] = numReqsTaught

	// Load number of comments on this user's pages
	var numCommentThreads int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM pagePairs
		WHERE type=?`, core.ParentPagePairType).Add(`
			AND parentId IN (
				SELECT  pageId
				FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
				WHERE pi.createdBy=?
					AND NOT pi.type=?)`, u.ID, core.CommentPageType).ToStatement(db).QueryRow()
	_, err = row.Scan(&numCommentThreads)
	if err != nil {
		return err
	}
	resultMap["numCommentThreads"] = numCommentThreads

	// Load number of replies to this user's comments
	var numReplies int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM pagePairs
		WHERE type=?`, core.ParentPagePairType).Add(`
			AND parentId IN (
				SELECT  pageId
				FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
				WHERE pi.createdBy=?
					AND pi.type=?)`, u.ID, core.CommentPageType).ToStatement(db).QueryRow()
	_, err = row.Scan(&numReplies)
	if err != nil {
		return err
	}
	resultMap["numReplies"] = numReplies

	// Load number of edits made.
	var numEdits int
	row = database.NewQuery(`
	    SELECT COUNT(*)
	    FROM pages
	    WHERE creatorId=?`, u.ID).ToStatement(db).QueryRow()
	_, err = row.Scan(&numEdits)
	if err != nil {
		return err
	}
	resultMap["numEdits"] = numEdits

	// Load number of pages edited.
	var numPagesEdited int
	row = database.NewQuery(`
	    SELECT COUNT(DISTINCT p.pageId)
		FROM `).AddPart(core.PageInfosTable(u)).Add(` AS pi
		JOIN pages AS p
		ON p.pageId=pi.pageId
		WHERE p.creatorId=?
			AND NOT pi.type=?`, u.ID, core.CommentPageType).ToStatement(db).QueryRow()
	_, err = row.Scan(&numPagesEdited)
	if err != nil {
		return err
	}
	resultMap["numPagesEdited"] = numPagesEdited

	// Load number of likes on my edits.
	var editLikes int
	row = database.NewQuery(`
	    SELECT COUNT(*)
	    FROM likes AS l
	    JOIN changeLogs AS cl
	    ON l.likeableId=cl.likeableId
	    JOIN pages AS p
	    ON cl.edit=p.edit AND cl.pageId=p.pageId
	    WHERE p.creatorId=?`, u.ID).ToStatement(db).QueryRow()
	_, err = row.Scan(&editLikes)
	if err != nil {
		return err
	}
	resultMap["editLikes"] = editLikes

	// Load number of answers.
	var numAnswers int
	row = database.NewQuery(`
	    SELECT COUNT(*)
	    FROM answers
	    WHERE userId=?`, u.ID).ToStatement(db).QueryRow()
	_, err = row.Scan(&numAnswers)
	if err != nil {
		return err
	}
	resultMap["numAnswers"] = numAnswers

	return nil
}

type dashboardListJSONData struct {
	NumToLoad int
}

type LoadFunction func(*core.CurrentUser, string, int, *database.DB, *core.CommonHandlerData, *core.PageLoadOptions) ([]string, error)

func DashboardListJSONHandler(params *pages.HandlerParams, loadFunction LoadFunction, listName string) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Decode data
	var data dashboardListJSONData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if data.NumToLoad <= 0 {
		data.NumToLoad = indexPanelLimit
	}

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{
		RedLinkCount: true,
	}).Add(core.TitlePlusLoadOptions)

	_, err = core.LoadAllDomainIDs(db, returnData.PageMap)
	if err != nil {
		return pages.Fail("Error while loading domain ids", err)
	}

	returnData.ResultMap[listName], err = loadFunction(u, params.PrivateGroupID, data.NumToLoad, db, returnData, pageOptions)
	if err != nil {
		return pages.Fail("error while loading "+listName, err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
