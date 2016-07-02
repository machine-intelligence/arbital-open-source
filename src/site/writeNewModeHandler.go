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
	NumPagesToLoad int
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
	if data.NumPagesToLoad <= 0 {
		data.NumPagesToLoad = DefaultModeRowCount
	}

	// Load redlinks in math
	returnData.ResultMap["redLinks"], err = loadRedLinkRows(db, returnData.User, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	return pages.Success(returnData)
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
	rows := database.NewQuery(`
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
		ORDER BY refCount + COALESCE(likeCount,0) DESC, groupedRedLinks.likeableId
		LIMIT ?`, limit).ToStatement(db).Query()
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
