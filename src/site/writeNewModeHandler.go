// Provide data for "write new" mode.
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	// If there is a draft with a given alias that's at most X days old, we won't
	// show the corresponding red link
	hideRedLinkIfDraftExistsDays = 3 // days
)

type writeNewModeData struct {
	IsFullPage bool
}

var writeNewModeHandler = siteHandler{
	URI:         "/json/writeNew/",
	HandlerFunc: writeNewModeHandlerFunc,
	Options:     pages.PageOptions{},
}

// Row to show a redLink
type RedLinkRow struct {
	core.Likeable
	Alias    string `json:"alias"`
	RefCount string `json:"refCount"`
}

type StubRow struct {
	PageID string `json:"pageId"`
}

func writeNewModeHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data writeNewModeData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	numPagesToLoad := 15
	if data.IsFullPage {
		numPagesToLoad = FullModeRowCount
	}

	// Load redlinks in math
	returnData.ResultMap["redLinks"], err = loadRedLinkRows(db, returnData.User, numPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load stubs in math
	returnData.ResultMap["stubs"], err = loadStubRows(db, returnData, numPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

func selectRandomNFrom(n int, query *database.QueryPart) *database.QueryPart {
	return database.NewQuery("SELECT * FROM (").AddPart(query).Add(") AS T ORDER BY RAND() LIMIT ?", n)
}

// Load pages that are linked to but don't exist
func loadRedLinkRows(db *database.DB, u *core.CurrentUser, limit int) ([]*RedLinkRow, error) {
	redLinks := make([]*RedLinkRow, 0)

	publishedPageIds := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Fields: []string{"pageId"},
	})
	// NOTE: keep in mind that multiple pages can have the same alias, as long as only one page is published
	publishedAndRecentAliases := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Unpublished: true,
		Fields:      []string{"alias"},
		WhereFilter: database.NewQuery(`currentEdit>0 OR DATEDIFF(NOW(),createdAt) <= ?`, hideRedLinkIfDraftExistsDays),
	})
	rows := selectRandomNFrom(limit, database.NewQuery(`
		SELECT childAlias,groupedRedLinks.likeableId,refCount
		FROM (
			SELECT l.childAlias,rl.likeableId,COUNT(*) AS refCount
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS mathPi
			JOIN pageDomainPairs AS pdp
			ON pdp.pageId=mathPi.pageId
				AND pdp.domainId=?`, core.MathDomainId).Add(`
			JOIN links AS l
			ON l.parentId=mathPi.pageId
				AND l.childAlias NOT IN`).AddPart(publishedPageIds).Add(`
				AND l.childAlias NOT IN`).AddPart(publishedAndRecentAliases).Add(`
			LEFT JOIN redLinks AS rl
			ON l.childAlias=rl.alias
			GROUP BY 1,2
		) AS groupedRedLinks
		LEFT JOIN (
			SELECT likeableId, SUM(value) AS likeCount, SUM(value < 0) AS hasAnyDownvotes
			FROM likes
			GROUP BY likeableId
		) as likeCounts
		ON groupedRedLinks.likeableId=likeCounts.likeableId
		WHERE !COALESCE(hasAnyDownvotes,0)
		ORDER BY COALESCE(likeCount,0) DESC, refCount DESC, groupedRedLinks.likeableId
		LIMIT ?`, 5*limit)).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var alias, refCount string
		var likeableId sql.NullInt64
		err := rows.Scan(&alias, &likeableId, &refCount)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		} else if core.IsIdValid(alias) {
			// Skip redlinks that are ids
			return nil
		}

		row := &RedLinkRow{
			Likeable: *core.NewLikeable(core.RedLinkLikeableType),
			Alias:    alias,
			RefCount: refCount,
		}
		if likeableId.Valid {
			row.LikeableId = likeableId.Int64
		}
		redLinks = append(redLinks, row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load red links: %v", err)
	}

	// Load likes
	likeablesMap := make(map[int64]*core.Likeable)
	for _, redLink := range redLinks {
		if redLink.LikeableId != 0 {
			likeablesMap[redLink.LikeableId] = &redLink.Likeable
		}
	}
	err = core.LoadLikes(db, u, likeablesMap, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load red link like count: %v", err)
	}

	return redLinks, nil
}

// Load pages that are marked as stubs
func loadStubRows(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*StubRow, error) {
	stubRows := make([]*StubRow, 0)
	rows := selectRandomNFrom(limit, database.NewQuery(`
		SELECT pi.pageId
		FROM`).AddPart(core.PageInfosTable(returnData.User)).Add(`AS pi
		JOIN pagePairs AS pp
		ON (pi.pageId=pp.childId)
		JOIN pageDomainPairs AS pdp
		ON (pi.pageId=pdp.pageId)
		LEFT JOIN likes AS l
		ON (pi.likeableId=l.likeableId)
		WHERE pp.parentId=?`, core.StubPageId).Add(`
			AND pdp.domainId=?`, core.MathDomainId).Add(`
			AND pi.lockedUntil < NOW()
		GROUP BY 1
		ORDER BY SUM(l.value) DESC
		LIMIT ?`, 5*limit)).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		err := rows.Scan(&pageID)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		stubRows = append(stubRows, &StubRow{PageID: pageID})
		core.AddPageIdToMap(pageID, returnData.PageMap)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load stub rows: %v", err)
	}
	return stubRows, nil
}
