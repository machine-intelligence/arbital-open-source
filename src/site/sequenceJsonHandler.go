// sequenceJsonHandler.go returns the sequence of pages needed for understanding a page
package site

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var sequenceHandler = siteHandler{
	URI:         "/json/sequence/",
	HandlerFunc: sequenceJsonHandler,
}

type sequenceJsonData struct {
	PageId int64 `json:",string"`
}

func sequenceJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	// Decode data
	var data sequenceJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}
	if data.PageId < 0 {
		return pages.HandlerBadRequestFail("Need a valid pageId", nil)
	}

	returnData := newHandlerData(true)
	returnData.User = u

	hasMastery := false
	row := database.NewQuery(`
		SELECT ifnull(max(has),false)
		FROM userMasteryPairs
		WHERE userId=?`, u.Id).Add(`AND masteryId=?`, data.PageId).ToStatement(db).QueryRow()
	_, err = row.Scan(&hasMastery)
	if err != nil {
		return pages.HandlerErrorFail("Error while checking if already knows", err)
	}
	childrenIds := make([]interface{}, 0)
	if !hasMastery {
		childrenIds = append(childrenIds, data.PageId)
	}

	// Track found pages so we can detect cycles
	requirementIds := make(map[int64]bool)
	requirementIds[data.PageId] = true

	// Recursively load all requirements
	for maxCount := 0; len(childrenIds) > 0 && maxCount < 20; maxCount++ {
		rows := database.NewQuery(`
			SELECT pp.parentId,pp.childId,mp.has
			FROM pagePairs AS pp
			LEFT JOIN userMasteryPairs AS mp
			ON (pp.parentId=mp.masteryId)
			WHERE mp.userId=?`, u.Id).Add(`AND pp.childId IN`).AddArgsGroup(childrenIds).Add(`
				AND pp.type=?`, core.RequirementPagePairType).ToStatement(db).Query()
		childrenIds := make([]interface{}, 0)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var parentId, childId int64
			var has sql.NullBool
			err := rows.Scan(&parentId, &childId, &has)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			if _, ok := requirementIds[parentId]; !ok {
				requirementIds[parentId] = true
				if has.Valid && !has.Bool {
					childrenIds = append(childrenIds, parentId)
				}
			} else {
				params.C.Warningf("Cycle detected with %d and %d", data.PageId, parentId)
			}
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while loading requirements", err)
		}
	}

	// Now load all the pages that teach the given requirements

	/*rows := database.NewQuery(`
		SELECT pp.parentId,pp.childId
		FROM pagePairs AS pp
		WHERE pp.parentId IN`).AddArgsGroup(requirementValues).Add(`
			AND pp.type=?`, core.SubjectPagePairType).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var parentId, childId int64
		err := rows.Scan(&parentId, &childId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		if _, ok := requirementIds[parentId]; !ok {
			requirementIds[parentId] = true
			if has.Valid && !has.Bool {
				childrenIds = append(childrenIds, parentId)
			}
		} else {
			params.C.Warningf("Cycle detected with %d and %d", data.PageId, parentId)
		}
		return nil
	})
	if err != nil {
		return pages.HandlerErrorFail("Error while loading requirements", err)
	}*/

	// Process all the found pages
	/*loadOptions := (&core.PageLoadOptions{
		Requirements: true,
	}).Add(core.TitlePlusLoadOptions)
	for pageId, _ := range foundPageIds {
		core.AddPageToMap(pageId, returnData.PageMap, loadOptions)
	}*/

	// Load pages.
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.toJson())
}
