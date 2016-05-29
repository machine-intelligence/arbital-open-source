// exploreHandler.go serves the data for /explore/ page
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	ExploreMaxPagesToLoad = 10000
)

type exploreJsonData struct {
	PageAlias string
}

var exploreHandler = siteHandler{
	URI:         "/json/explore/",
	HandlerFunc: exploreJsonHandler,
}

func exploreJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(params.U).SetResetEverything()

	// Decode data
	var data exploreJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageId, ok, err := core.LoadAliasToPageId(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	} else if !ok {
		return pages.Fail("No such page found", err)
	}
	returnData.ResultMap["pageId"] = pageId

	loadOptions := (&core.PageLoadOptions{
		SubpageCounts:    true,
		AnswerCounts:     true,
		Children:         true,
		HasGrandChildren: true,
		Tags:             true,
		Lenses:           true,
		ViewCount:        true,
		IsSubscribed:     true,
		RedLinkCount:     true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(pageId, returnData.PageMap, loadOptions)

	// Load all children of the pageId, then load all the grand-children, etc. recursively
	parentIdsToProcess := []string{pageId}
	for len(parentIdsToProcess) > 0 {
		rows := database.NewQuery(`
			SELECT pp.childId
			FROM pagePairs AS pp
			JOIN`).AddPart(core.PageInfosTable(u)).Add(`AS pi
			ON (pi.pageId=pp.childId)
			WHERE pp.type=?`, core.ParentPagePairType).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
				AND pp.parentId IN`).AddArgsGroupStr(parentIdsToProcess).ToStatement(db).Query()
		parentIdsToProcess = make([]string, 0)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId string
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			_, ok := returnData.PageMap[pageId]
			if !ok && len(returnData.PageMap) < ExploreMaxPagesToLoad {
				parentIdsToProcess = append(parentIdsToProcess, pageId)
				core.AddPageToMap(pageId, returnData.PageMap, loadOptions)
			}
			return nil
		})
		if err != nil {
			return pages.Fail("Couldn't load children", err)
		}
		if len(returnData.PageMap) >= ExploreMaxPagesToLoad {
			break
		}
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
