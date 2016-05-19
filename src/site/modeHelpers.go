// Contains helpers for various modes.
package site

import (
	"fmt"
	"sort"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	DefaultModeRowCount = 25

	PageModeRowType       = "page"
	CommentModeRowType    = "comment"
	MarkModeRowType       = "mark"
	QueryModeRowType      = "query"
	LikesModeRowType      = "likes"
	ReqsTaughtModeRowType = "reqsTaught"
)

type modeRowData struct {
	RowType      string `json:"rowType"`
	ActivityDate string `json:"activityDate"`
}

type modeRow interface {
	GetActivityDate() string
}

func (row *modeRowData) GetActivityDate() string {
	return row.ActivityDate
}

// Row for a new comment or reply
type commentModeRow struct {
	modeRowData
	CommentId string `json:"commentId"`
}

// Row for some kind of mark event
type markModeRow struct {
	modeRowData
	MarkId string `json:"markId"`
}

// Row to show which users like a page
type likesModeRow struct {
	modeRowData
	PageId    string          `json:"pageId"`
	ChangeLog *core.ChangeLog `json:"changeLog"` // Optional changeLog.
	UserIds   []string        `json:"userIds"`
}

// Row to show which users learned some requisites
type reqsTaughtModeRow struct {
	modeRowData
	PageId       string   `json:"pageId"`
	UserIds      []string `json:"userIds"`
	RequisiteIds []string `json:"requisiteIds"`
}

// Row to show a page
type pageModeRow struct {
	modeRowData
	PageId string `json:"pageId"`
}

type ModeRows []modeRow

func (a ModeRows) Len() int           { return len(a) }
func (a ModeRows) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ModeRows) Less(i, j int) bool { return a[i].GetActivityDate() > a[j].GetActivityDate() }

// Take a list of ModeRows, combine them into one list, sorted by date, and then
// return at most "limit" most recent rows.
func combineModeRows(limit int, listOfRows ...ModeRows) ModeRows {
	allRows := make(ModeRows, 0)
	for _, rows := range listOfRows {
		allRows = append(allRows, rows...)
	}
	sort.Sort(allRows)
	if len(allRows) > limit {
		allRows = allRows[:limit]
	}
	return allRows
}

// Load all the comments.
func loadCommentModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	modeRows := make(ModeRows, 0)
	parentPageOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)
	childPageOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.TitlePlusLoadOptions)
	rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId,pi.createdAt
		FROM`).AddPart(core.PageInfosTable(returnData.User)).Add(`AS pi
		JOIN pagePairs AS pp
		ON (pp.childId=pi.pageId)
		JOIN subscriptions AS s
		ON (pp.parentId=s.toId)
		WHERE s.userId=?`, returnData.User.Id).Add(`
			AND pi.createdBy!=?`, returnData.User.Id).Add(`
			AND pi.type=?`, core.CommentPageType).Add(`
			AND NOT pi.isEditorComment
		GROUP BY pp.childId
		ORDER BY pi.createdAt DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId, activityDate string
		err := rows.Scan(&parentId, &childId, &activityDate)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		modeRows = append(modeRows, &commentModeRow{
			modeRowData: modeRowData{RowType: CommentModeRowType, ActivityDate: activityDate},
			CommentId:   childId,
		})
		core.AddPageToMap(parentId, returnData.PageMap, parentPageOptions)
		core.AddPageToMap(childId, returnData.PageMap, childPageOptions)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error reading rows: %v", err)
	}

	return modeRows, nil
}

// Load all the marks.
func loadMarkModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	modeRows := make(ModeRows, 0)
	rows := database.NewQuery(`
		SELECT id,type,IF(answeredAt="",resolvedAt,answeredAt)
		FROM marks
		WHERE creatorId=?`, returnData.User.Id).Add(`
			AND resolvedAt!=""
		ORDER BY 2 DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var markId, markType, activityDate string
		err := rows.Scan(&markId, &markType, &activityDate)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		modeRowType := MarkModeRowType
		if markType == core.QueryMarkType {
			modeRowType = QueryModeRowType
		}
		modeRows = append(modeRows, &markModeRow{
			modeRowData: modeRowData{RowType: modeRowType, ActivityDate: activityDate},
			MarkId:      markId,
		})
		returnData.MarkMap[markId] = &core.Mark{Id: markId}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error reading rows: %v", err)
	}

	return modeRows, nil
}

func loadLikesModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	hedonsRowMap := make(map[string]*likesModeRow)

	rows := database.NewQuery(`
		SELECT u.id,pi.pageId,pi.type,l.updatedAt
		FROM `).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
		JOIN likes AS l
		ON pi.likeableId=l.likeableId
		JOIN users AS u
		ON l.userId=u.id
		WHERE pi.createdBy=?`, returnData.User.Id).Add(`
			AND l.userId!=?`, returnData.User.Id).Add(`
			AND l.value=1 
		ORDER BY l.updatedAt DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likerId, pageId, pageType, updatedAt string

		err := rows.Scan(&likerId, &pageId, &pageType, &updatedAt)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		row, ok := hedonsRowMap[pageId]
		if !ok {
			row = &likesModeRow{
				modeRowData: modeRowData{RowType: LikesModeRowType, ActivityDate: updatedAt},
				PageId:      pageId,
				UserIds:     make([]string, 0),
			}
			hedonsRowMap[pageId] = row

			loadOptions := core.TitlePlusLoadOptions
			if pageType == core.CommentPageType {
				loadOptions.Add(&core.PageLoadOptions{Parents: true})
			}
			core.AddPageToMap(pageId, returnData.PageMap, loadOptions)
		}

		core.AddUserIdToMap(likerId, returnData.UserMap)
		row.UserIds = append(row.UserIds, likerId)
		return nil
	})
	if err != nil {
		return nil, err
	}

	modeRows := make(ModeRows, 0)
	for _, row := range hedonsRowMap {
		modeRows = append(modeRows, row)
	}

	return modeRows, nil
}

func loadChangeLikesModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	hedonsRowMap := make(map[int]*likesModeRow, 0)

	rows := database.NewQuery(`
		SELECT l.userId,cl.pageId,l.updatedAt,cl.id,cl.pageId,cl.type,cl.oldSettingsValue,cl.newSettingsValue,cl.edit
		FROM likes as l
		JOIN changeLogs as cl
		ON cl.likeableId=l.likeableId
		WHERE cl.userId=?`, returnData.User.Id).Add(`
			AND l.value=1 AND l.userId!=?`, returnData.User.Id).Add(`
			AND cl.type=?`, core.NewEditChangeLog).Add(`
		ORDER BY l.updatedAt DESC`).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var likerId, pageId, updatedAt string
		var changeLog core.ChangeLog

		err := rows.Scan(&likerId, &pageId, &updatedAt,
			&changeLog.Id, &changeLog.PageId, &changeLog.Type, &changeLog.OldSettingsValue,
			&changeLog.NewSettingsValue, &changeLog.Edit)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		row, ok := hedonsRowMap[changeLog.Id]
		if !ok {
			row = &likesModeRow{
				modeRowData: modeRowData{RowType: LikesModeRowType, ActivityDate: updatedAt},
				PageId:      pageId,
				ChangeLog:   &changeLog,
				UserIds:     make([]string, 0),
			}
			hedonsRowMap[changeLog.Id] = row
		}

		core.AddPageToMap(pageId, returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddUserIdToMap(likerId, returnData.UserMap)
		row.UserIds = append(row.UserIds, likerId)
		return nil
	})
	if err != nil {
		return nil, err
	}

	hedonsRows := make(ModeRows, 0)
	for _, row := range hedonsRowMap {
		hedonsRows = append(hedonsRows, row)
	}

	return hedonsRows, nil
}

// Load all the requisites taught by this user.
func loadReqsTaughtModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	hedonsRowMap := make(map[string]*reqsTaughtModeRow)

	rows := database.NewQuery(`
		SELECT u.id,pi.pageId,ump.masteryId,ump.updatedAt
		FROM userMasteryPairs AS ump
		JOIN `).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
		ON ump.taughtBy=pi.pageId
		JOIN users AS u
		ON ump.userId=u.id
		WHERE pi.createdBy=?`, returnData.User.Id).Add(`
			AND ump.has=1 AND ump.userId!=?`, returnData.User.Id).Add(`
		ORDER BY ump.updatedAt DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var learnerId, taughtById, masteryId, updatedAt string

		err := rows.Scan(&learnerId, &taughtById, &masteryId, &updatedAt)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}

		row, ok := hedonsRowMap[taughtById]
		if !ok {
			row = &reqsTaughtModeRow{
				modeRowData:  modeRowData{RowType: ReqsTaughtModeRowType, ActivityDate: updatedAt},
				PageId:       taughtById,
				UserIds:      make([]string, 0),
				RequisiteIds: make([]string, 0),
			}
			hedonsRowMap[taughtById] = row
		}

		core.AddPageToMap(taughtById, returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddPageToMap(masteryId, returnData.PageMap, core.TitlePlusLoadOptions)
		core.AddUserIdToMap(learnerId, returnData.UserMap)
		if !core.IsStringInList(learnerId, row.UserIds) {
			row.UserIds = append(row.UserIds, learnerId)
		}
		if !core.IsStringInList(masteryId, row.RequisiteIds) {
			row.RequisiteIds = append(row.RequisiteIds, masteryId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	modeRows := make(ModeRows, 0)
	for _, row := range hedonsRowMap {
		modeRows = append(modeRows, row)
	}

	return modeRows, nil
}

func loadHotPagesModeRows(db *database.DB, returnData *core.CommonHandlerData, limit int) (ModeRows, error) {
	modeRows := make(ModeRows, 0)
	pageLoadOptions := (&core.PageLoadOptions{
		SubpageCounts: true,
		AnswerCounts:  true,
	}).Add(core.TitlePlusLoadOptions)

	rows := database.NewQuery(`
		SELECT pageId,createdAt
		FROM`).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
		WHERE pi.type IN (?,?,?,?)`, core.WikiPageType, core.LensPageType, core.DomainPageType, core.QuestionPageType).Add(`
		ORDER BY createdAt DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId, activityDate string
		err := rows.Scan(&pageId, &activityDate)
		if err != nil {
			return fmt.Errorf("failed to scan a pageId: %v", err)
		}
		row := &pageModeRow{
			modeRowData: modeRowData{RowType: PageModeRowType, ActivityDate: activityDate},
			PageId:      pageId,
		}
		modeRows = append(modeRows, row)

		core.AddPageToMap(pageId, returnData.PageMap, pageLoadOptions)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return modeRows, nil
}
