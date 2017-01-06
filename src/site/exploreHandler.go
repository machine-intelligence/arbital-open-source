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

type exploreJSONData struct {
	PageAlias string
}

var exploreHandler = siteHandler{
	URI:         "/json/explore/",
	HandlerFunc: exploreJSONHandler,
}

func exploreJSONHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(params.U).SetResetEverything()

	// Decode data
	var data exploreJSONData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Get actual page id
	pageID, ok, err := core.LoadAliasToPageID(db, u, data.PageAlias)
	if err != nil {
		return pages.Fail("Couldn't convert alias", err)
	} else if !ok {
		return pages.Fail("No such page found", err)
	}
	returnData.ResultMap["pageId"] = pageID

	loadOptions := (&core.PageLoadOptions{
		SubpageCounts:    true,
		AnswerCounts:     true,
		Children:         true,
		HasGrandChildren: true,
		Tags:             true,
		Lenses:           true,
		IsSubscribed:     true,
		RedLinkCount:     true,
	}).Add(core.TitlePlusLoadOptions)
	core.AddPageToMap(pageID, returnData.PageMap, loadOptions)

	// Load all children of the pageId, then load all the grand-children, etc. recursively
	parentIdsToProcess := []string{pageID}
	for len(parentIdsToProcess) > 0 {
		rows := database.NewQuery(`
			SELECT pp.childId
			FROM pagePairs AS pp
			JOIN pageInfos AS pi
			ON (pi.pageId=pp.childId)
			WHERE pp.type=?`, core.ParentPagePairType).Add(`
				AND pi.type!=?`, core.CommentPageType).Add(`
				AND pp.parentId IN`).AddArgsGroupStr(parentIdsToProcess).Add(`
				AND`).AddPart(core.PageInfosFilter(u)).ToStatement(db).Query()
		parentIdsToProcess = make([]string, 0)
		err := rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageID string
			err := rows.Scan(&pageID)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			_, ok := returnData.PageMap[pageID]
			if !ok && len(returnData.PageMap) < ExploreMaxPagesToLoad {
				parentIdsToProcess = append(parentIdsToProcess, pageID)
				core.AddPageToMap(pageID, returnData.PageMap, loadOptions)
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
