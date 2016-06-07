// Provide data for "write new" mode.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type writeNewModeData struct {
	NumPagesToLoad int
}

var writeNewModeHandler = siteHandler{
	URI:         "/json/writeNew/",
	HandlerFunc: writeNewModeHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// Row to show a redLink
type RedLinkRow struct {
	RedLinkAlias string `json:"redLinkAlias"`
	RedLinkCount string `json:"redLinkCount"`
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
	returnData.ResultMap["redLinks"], err = loadRedLinkRows(db, returnData, data.NumPagesToLoad)
	if err != nil {
		return pages.Fail("Error loading drafts", err)
	}

	// Load pages
	core.AddPageIdToMap("3hs", returnData.PageMap)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}

// Load pages that are linked to but don't exist
func loadRedLinkRows(db *database.DB, returnData *core.CommonHandlerData, limit int) ([]*RedLinkRow, error) {
	redLinks := make([]*RedLinkRow, 0)

	rows := database.NewQuery(`
		SELECT l.childAlias,SUM(ISNULL(linkedPi.pageId)) AS count
		FROM pageInfos as mathPi
		JOIN pageDomainPairs as pdp
		ON pdp.pageId=mathPi.pageId
			AND pdp.domainId=?`, MathDomainId).Add(`
		JOIN links as l
		ON l.parentId=mathPi.pageId
		LEFT JOIN pageInfos as linkedPi
		ON (l.childAlias=linkedPi.pageId OR l.childAlias=linkedPi.alias)
		GROUP BY l.childAlias
		ORDER BY count DESC
		LIMIT ?`, limit).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		var redLinkAlias, redLinkCount string
		err := rows.Scan(&redLinkAlias, &redLinkCount)
		if err != nil {
			return fmt.Errorf("failed to scan: %v", err)
		}
		row := &RedLinkRow{
			RedLinkAlias: redLinkAlias,
			RedLinkCount: redLinkCount,
		}
		redLinks = append(redLinks, row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return redLinks, nil
}
