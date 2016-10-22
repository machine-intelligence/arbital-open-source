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

	// varietyDepth represents how far down we sample the top new pages.
	// When loading N pages, we will randomly sample them from the top
	// varietyDepth*N results.
	varietyDepth = 5
)

type writeNewModeData struct {
	IsFullPage bool
}

var writeNewModeHandler = siteHandler{
	URI:         "/json/writeNew/",
	HandlerFunc: writeNewModeHandlerFunc,
	Options:     pages.PageOptions{},
}

type RedLinkRow struct {
	core.Likeable
	Alias           string   `json:"alias"`
	LinkedByPageIDs []string `json:"linkedByPageIds"`
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
	returnData.ResultMap["redLinks"], err = loadRedLinkRows(db, returnData, numPagesToLoad, "")
	if err != nil {
		return pages.Fail("Error loading red link rows", err)
	}

	// Load contentRequests in math
	returnData.ResultMap["contentRequests"], err = loadContentRequests(db, returnData, numPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading content requests", err)
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
func loadRedLinkRows(db *database.DB, returnData *core.CommonHandlerData, limit int, optionalPageAlias string) ([]*RedLinkRow, error) {
	u := returnData.User
	redLinks := []*RedLinkRow{}

	publishedPageIDs := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Fields: []string{"pageId"},
	})
	// NOTE: keep in mind that multiple pages can have the same alias, as long as only one page is published
	publishedAndRecentAliases := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Unpublished: true,
		Fields:      []string{"alias"},
		WhereFilter: database.NewQuery(`currentEdit > 0 OR DATEDIFF(NOW(),createdAt) <= ?`, hideRedLinkIfDraftExistsDays),
	})
	optionalPageAliasConstraint := database.NewQuery(``)
	if optionalPageAlias != "" {
		optionalPageAliasConstraint.Add(`WHERE l.childAlias=?`, optionalPageAlias)
	}
	rows := selectRandomNFrom(limit, database.NewQuery(`
		SELECT childAlias,groupedRedLinks.likeableId
		FROM (
			SELECT l.childAlias,rl.likeableId,COUNT(*) AS refCount
			FROM`).AddPart(core.PageInfosTable(u)).Add(`AS mathPi
			JOIN pageDomainPairs AS pdp
			ON pdp.pageId=mathPi.pageId
				AND pdp.domainId=?`, core.MathDomainID).Add(`
			JOIN links AS l
			ON l.parentId=mathPi.pageId
				AND l.childAlias NOT IN`).AddPart(publishedPageIDs).Add(`
				AND l.childAlias NOT IN`).AddPart(publishedAndRecentAliases).Add(`
			LEFT JOIN redLinks AS rl
			ON l.childAlias=rl.alias`).AddPart(optionalPageAliasConstraint).Add(`
			GROUP BY 1,2
		) AS groupedRedLinks
		LEFT JOIN (
			SELECT likeableId, SUM(value) AS likeCount, SUM(value < 0) AS hasAnyDownvotes
			FROM likes
			GROUP BY likeableId
		) AS likeCounts
		ON groupedRedLinks.likeableId=likeCounts.likeableId
		WHERE !COALESCE(hasAnyDownvotes,0)
		ORDER BY COALESCE(likeCount,0) DESC, refCount DESC, groupedRedLinks.likeableId
		LIMIT ?`, varietyDepth*limit)).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var alias string
		var likeableID sql.NullInt64
		err := rows.Scan(&alias, &likeableID)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		} else if core.IsIDValid(alias) {
			// Skip redlinks that are ids
			return nil
		}

		row := &RedLinkRow{
			Likeable: *core.NewLikeable(core.RedLinkLikeableType),
			Alias:    alias,
		}
		if likeableID.Valid {
			row.LikeableID = likeableID.Int64
		}
		redLinks = append(redLinks, row)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Couldn't load red links: %v", err)
	}

	// Load likes
	{
		likeablesMap := make(map[int64]*core.Likeable)
		for _, redLink := range redLinks {
			if redLink.LikeableID != 0 {
				likeablesMap[redLink.LikeableID] = &redLink.Likeable
			}
		}
		err := core.LoadLikes(db, u, likeablesMap, likeablesMap, returnData.UserMap)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load red link like count: %v", err)
		}
	}

	// Load related pages
	{
		aliases := make([]string, len(redLinks))
		for i := range aliases {
			aliases[i] = redLinks[i].Alias
		}
		related, err := loadRelationships(db, aliases, returnData, true)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load relationships: %v", err)
		}
		for _, row := range redLinks {
			row.LinkedByPageIDs = related[row.Alias]
		}
	}

	return redLinks, nil
}

func loadContentRequests(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*core.ContentRequest, error) {
	contentRequests := []*core.ContentRequest{}

	// Load content requests, sorted by likes
	{
		rows := database.NewQuery(`
			SELECT cr.id, cr.pageId, cr.type, cr.likeableId, cr.createdAt
			FROM contentRequests AS cr
			LEFT JOIN (
				SELECT likeableId, SUM(value) AS likeCount, SUM(value < 0) AS hasAnyDownvotes
				FROM likes
				GROUP BY likeableId
			) AS likeCounts
			ON (cr.likeableId=likeCounts.likeableId)
			JOIN `).AddPart(core.PageInfosTable(returnData.User)).Add(` AS pi
			ON cr.pageId=pi.pageId
			/* Having no dislikes */
			WHERE !COALESCE(hasAnyDownvotes,0)
				AND NOT cr.type IN (?,?)`, string(core.SlowDown), string(core.SpeedUp)).Add(`
			GROUP BY cr.id
			ORDER BY COALESCE(likeCount, 0) DESC
			LIMIT ?`, limit).ToStatement(db).Query()
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			cr := core.NewContentRequest()
			err := rows.Scan(&cr.ID, &cr.PageID, &cr.RequestType, &cr.LikeableID, &cr.CreatedAt)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}

			contentRequests = append(contentRequests, cr)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Load likes for the content requests
	{
		likeablesMap := make(map[int64]*core.Likeable)
		for _, cr := range contentRequests {
			l := &cr.Likeable
			if l.LikeableID != 0 {
				likeablesMap[l.LikeableID] = l
			}
		}
		err := core.LoadLikes(db, returnData.User, likeablesMap, likeablesMap, returnData.UserMap)
		if err != nil {
			return nil, err
		}
	}

	// Add content request pages to the page map (so we can show their titles, etc)
	for _, cr := range contentRequests {
		core.AddPageIDToMap(cr.PageID, returnData.PageMap)
	}

	return contentRequests, nil
}
