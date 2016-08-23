// Provide data for a project page.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type projectParams struct {
}

var projectHandler = siteHandler{
	URI:         "/json/project/",
	HandlerFunc: projectHandlerFunc,
	Options:     pages.PageOptions{},
}

type ProjectData struct {
	// All red aliases for this project
	AliasRows []*ProjectAliasRow `json:"aliasRows"`
	// All pages for this project
	PageIDs []string `json:"pageIds"`
}

type ProjectAliasRow struct {
	core.Likeable
	Alias string `json:"alias"`
}

func projectHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u)

	// Decode data
	var data projectParams
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	numPagesToLoad := 100
	var projectData ProjectData

	// Load redlinks in math
	projectData.AliasRows, err = loadProjectRedAliasRows(db, returnData, numPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading red aliases", err)
	}

	// Load stubs in math
	projectData.PageIDs, err = loadProjectPageIDs(db, returnData, numPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading project pages", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["projectData"] = projectData
	return pages.Success(returnData)
}

// Load pages that will be part of the project but don't exist yet and need to be created
func loadProjectRedAliasRows(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*ProjectAliasRow, error) {
	u := returnData.User
	redLinks := []*ProjectAliasRow{}

	publishedPageIDs := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Fields: []string{"pageId"},
	})
	// NOTE: keep in mind that multiple pages can have the same alias, as long as only one page is published
	publishedAndRecentAliases := core.PageInfosTableWithOptions(u, &core.PageInfosOptions{
		Unpublished: true,
		Fields:      []string{"alias"},
		WhereFilter: database.NewQuery(`currentEdit>0 OR DATEDIFF(NOW(),createdAt) <= ?`, hideRedLinkIfDraftExistsDays),
	})
	rows := database.NewQuery(`
		SELECT l.childAlias
		FROM links AS l
		LEFT JOIN redLinks AS rl
		ON l.childAlias=rl.alias
		WHERE l.parentId=/*'5wy'*/'5mv'
		AND l.childAlias NOT IN`).AddPart(publishedPageIDs).Add(`
		AND l.childAlias NOT IN`).AddPart(publishedAndRecentAliases).Add(`
		GROUP BY 1
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var alias string
		err := rows.Scan(&alias)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		} else if core.IsIDValid(alias) {
			// Skip redlinks that are ids
			return nil
		}

		row := &ProjectAliasRow{
			Likeable: *core.NewLikeable(core.RedLinkLikeableType),
			Alias:    alias,
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

	return redLinks, nil
}

// Load pages that are part of the project
func loadProjectPageIDs(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]string, error) {
	pageIDs := []string{}
	rows := database.NewQuery(`
		SELECT pi.pageId
		FROM`).AddPart(core.PageInfosTable(returnData.User)).Add(`AS pi
		JOIN links AS l
		ON (l.childAlias=pi.pageId AND l.parentId=/*'5wy'*/'5mv')
		GROUP BY 1
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		err := rows.Scan(&pageID)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		pageIDs = append(pageIDs, pageID)
		core.AddPageToMap(pageID, returnData.PageMap, &core.PageLoadOptions{
			Tags:       true,
			Likes:      true,
			Text:       true,
			ChangeLogs: true,
		})
		return nil
	})
	return pageIDs, err
}
