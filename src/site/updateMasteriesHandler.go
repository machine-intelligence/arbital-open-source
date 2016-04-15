// updateMasteries.go handles request to add and/or delete masteries
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateMasteries contains the data we get in the request.
type updateMasteries struct {
	RemoveMasteries []string
	WantsMasteries  []string
	AddMasteries    []string
	// The id of the page that taught these masteries, if any
	TaughtBy string
	// If true, compute which pages the user can now read
	ComputeUnlocked bool
}

var updateMasteriesHandler = siteHandler{
	URI:         "/updateMasteries/",
	HandlerFunc: updateMasteriesHandlerFunc,
	Options: pages.PageOptions{
		LoadUserTrust: true,
	},
}

func updateMasteriesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data updateMasteries
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	return updateMasteriesInternalHandlerFunc(params, &data)
}

func updateMasteriesInternalHandlerFunc(params *pages.HandlerParams, data *updateMasteries) *pages.Result {
	db := params.DB
	u := params.U

	userId := u.GetSomeId()
	if userId == "" {
		return pages.HandlerBadRequestFail("No user id or session id", nil)
	}

	allMasteries := append(append(data.AddMasteries, data.RemoveMasteries...), data.WantsMasteries...)
	aliasMap, err := core.LoadAliasToPageIdMap(db, allMasteries)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't translate aliases to ids", err)
	}

	subjectIds := make(map[string]bool)
	if data.TaughtBy != "" {
		rows := db.NewStatement(`
			SELECT parentId from pagePairs
			WHERE childId=? AND type=?`).Query(data.TaughtBy, core.SubjectPagePairType)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var subjectId string
			err := rows.Scan(&subjectId)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			subjectIds[subjectId] = true
			return nil
		})
	}

	candidateIds := make([]string, 0)
	if data.ComputeUnlocked && len(data.AddMasteries) > 0 {
		// Compute all the pages that rely on at least one of these masteries, that the user can't yet understand
		rows := database.NewQuery(`
			SELECT pp1.childId
			FROM pagePairs AS pp1
			JOIN pagePairs AS pp2
			ON (pp1.childId=pp2.childId)
			JOIN userMasteryPairs AS mp
			ON (pp2.parentId=mp.masteryId AND mp.userId=?)`, userId).Add(`
			WHERE pp1.parentId IN`).AddArgsGroupStr(data.AddMasteries).Add(`
				AND pp1.type=?`, core.RequirementPagePairType).Add(`
				AND pp2.type=?`, core.RequirementPagePairType).Add(`
				AND NOT mp.has
			GROUP BY 1`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageId string
			err := rows.Scan(&pageId)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			candidateIds = append(candidateIds, pageId)
			return nil
		})
		if err != nil {
			return pages.HandlerErrorFail("Error while loading potential unlocked ids", err)
		}
	}

	_, err = db.Transaction(func(tx *database.Tx) (string, error) {
		snapshotId, err := InsertUserTrustSnapshots(tx, u, data.TaughtBy)
		if err != nil {
			return "Couldn't insert userTrustSnapshot", err
		}

		hashmaps := make(database.InsertMaps, 0)
		for _, masteryAlias := range data.RemoveMasteries {
			if masteryId, ok := aliasMap[masteryAlias]; ok {
				hashmap := getHashmapForMasteryInsert(masteryId, userId, false, false, "", snapshotId)
				hashmaps = append(hashmaps, hashmap)
			}
		}
		for _, masteryAlias := range data.WantsMasteries {
			if masteryId, ok := aliasMap[masteryAlias]; ok {
				hashmap := getHashmapForMasteryInsert(masteryId, userId, false, true, "", snapshotId)
				hashmaps = append(hashmaps, hashmap)
			}
		}
		for _, masteryAlias := range data.AddMasteries {
			if masteryId, ok := aliasMap[masteryAlias]; ok {
				var taughtBy = ""
				if _, ok := subjectIds[masteryId]; ok {
					taughtBy = data.TaughtBy
				}
				hashmap := getHashmapForMasteryInsert(masteryId, userId, true, false, taughtBy, snapshotId)
				hashmaps = append(hashmaps, hashmap)
			}
		}

		statement := tx.DB.NewMultipleInsertStatement("userMasteryPairs", hashmaps, "has", "wants", "updatedAt", "taughtBy", "userTrustSnapshotId")
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return "Failed to insert masteries", err
		}
		return "", nil
	})

	if err != nil {
		return pages.HandlerErrorFail("Couldn't update masteries", err)
	}

	if len(candidateIds) <= 0 {
		return pages.StatusOK(nil)
	}

	unlockedIds := make([]string, 0)
	returnData := core.NewHandlerData(params.U, false)

	// For the previously computed candidates, check if the user can now understand them
	rows := database.NewQuery(`
		SELECT pp.childId
		FROM pagePairs AS pp
		LEFT JOIN userMasteryPairs AS mp
		ON (pp.parentId=mp.masteryId AND mp.userId=?)`, userId).Add(`
		WHERE pp.childId IN`).AddArgsGroupStr(candidateIds).Add(`
			AND pp.type=?`, core.RequirementPagePairType).Add(`
		GROUP BY 1
		HAVING SUM(1)<=SUM(mp.has)
		LIMIT 5`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageId string
		err := rows.Scan(&pageId)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		unlockedIds = append(unlockedIds, pageId)
		core.AddPageToMap(pageId, returnData.PageMap, core.TitlePlusLoadOptions)
		return nil
	})
	if err != nil {
		return pages.HandlerErrorFail("Error while loading unlocked ids", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["unlockedIds"] = unlockedIds
	return pages.StatusOK(returnData)

}

func getHashmapForMasteryInsert(masteryId string, userId string, has bool, wants bool, taughtBy string, snapshotId int64) database.InsertMap {
	hashmap := make(database.InsertMap)
	hashmap["masteryId"] = masteryId
	hashmap["userId"] = userId
	hashmap["has"] = has
	hashmap["wants"] = wants
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	hashmap["taughtBy"] = taughtBy
	hashmap["userTrustSnapshotId"] = snapshotId

	return hashmap
}
