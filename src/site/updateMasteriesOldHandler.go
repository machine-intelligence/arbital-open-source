// updateMasteriesOld.go handles request to add and/or delete masteries

package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// updateMasteriesOld contains the data we get in the request.
type updateMasteriesOld struct {
	RemoveMasteries []string
	WantsMasteries  []string
	AddMasteries    []string
	// The id of the page that taught these masteries, if any
	TaughtBy string
	// If true, compute which pages the user can now read
	ComputeUnlocked bool
}

var updateMasteriesOldHandler = siteHandler{
	URI:         "/updateMasteriesOld/",
	HandlerFunc: updateMasteriesOldHandlerFunc,
	Options:     pages.PageOptions{},
}

func updateMasteriesOldHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data updateMasteriesOld
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	return updateMasteriesInternalOldHandlerFunc(params, &data)
}

func updateMasteriesInternalOldHandlerFunc(params *pages.HandlerParams, data *updateMasteriesOld) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	userID := u.GetSomeID()
	if userID == "" {
		return pages.Fail("No user id or session id", nil).Status(http.StatusBadRequest)
	}

	// Translate all mastery aliases to ids
	allMasteries := append(append(data.AddMasteries, data.RemoveMasteries...), data.WantsMasteries...)
	aliasMap, err := core.LoadAliasToPageIDMap(db, u, allMasteries)
	if err != nil {
		return pages.Fail("Couldn't translate aliases to ids", err)
	}

	subjectIDs := make(map[string]bool)
	if data.TaughtBy != "" {
		rows := db.NewStatement(`
			SELECT parentId
			FROM pagePairs
			WHERE childId=? AND type=?`).Query(data.TaughtBy, core.SubjectPagePairType)
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var subjectID string
			err := rows.Scan(&subjectID)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			subjectIDs[subjectID] = true
			return nil
		})
	}

	candidateIDs := make([]string, 0)
	if data.ComputeUnlocked && len(data.AddMasteries) > 0 {
		// Compute all the pages that rely on at least one of these masteries, that the user can't yet understand
		rows := database.NewQuery(`
			SELECT pp1.childId
			FROM pagePairs AS pp1
			JOIN pagePairs AS pp2
			ON (pp1.childId=pp2.childId)
			JOIN userMasteryPairs AS mp
			ON (pp2.parentId=mp.masteryId AND mp.userId=?)`, userID).Add(`
			WHERE pp1.parentId IN`).AddArgsGroupStr(data.AddMasteries).Add(`
				AND pp1.type=?`, core.RequirementPagePairType).Add(`
				AND pp2.type=?`, core.RequirementPagePairType).Add(`
				AND NOT mp.has
			GROUP BY 1`).ToStatement(db).Query()
		err = rows.Process(func(db *database.DB, rows *database.Rows) error {
			var pageID string
			err := rows.Scan(&pageID)
			if err != nil {
				return fmt.Errorf("Failed to scan: %v", err)
			}
			candidateIDs = append(candidateIDs, pageID)
			return nil
		})
		if err != nil {
			return pages.Fail("Error while loading potential unlocked ids", err)
		}
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		hashmaps := make(database.InsertMaps, 0)
		for _, masteryAlias := range data.RemoveMasteries {
			if masteryID, ok := aliasMap[masteryAlias]; ok {
				hashmap := getHashmapForMasteryInsert(masteryID, userID, false, false, "")
				hashmaps = append(hashmaps, hashmap)
			}
		}
		for _, masteryAlias := range data.WantsMasteries {
			if masteryID, ok := aliasMap[masteryAlias]; ok {
				hashmap := getHashmapForMasteryInsert(masteryID, userID, false, true, "")
				hashmaps = append(hashmaps, hashmap)
			}
		}
		for _, masteryAlias := range data.AddMasteries {
			if masteryID, ok := aliasMap[masteryAlias]; ok {
				var taughtBy = ""
				if _, ok := subjectIDs[masteryID]; ok {
					taughtBy = data.TaughtBy
				}
				hashmap := getHashmapForMasteryInsert(masteryID, userID, true, false, taughtBy)
				hashmaps = append(hashmaps, hashmap)
			}
		}

		if len(hashmaps) > 0 {
			statement := tx.DB.NewMultipleInsertStatement("userMasteryPairs", hashmaps, "has", "wants", "updatedAt", "taughtBy")
			if _, err := statement.WithTx(tx).Exec(); err != nil {
				return sessions.NewError("Failed to insert masteries", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	if len(candidateIDs) <= 0 {
		return pages.Success(nil)
	}

	// For the previously computed candidates, check if the user can now understand them
	unlockedIDs := make([]string, 0)
	rows := database.NewQuery(`
		SELECT pp.childId
		FROM pagePairs AS pp
		LEFT JOIN userMasteryPairs AS mp
		ON (pp.parentId=mp.masteryId AND mp.userId=?)`, userID).Add(`
		WHERE pp.childId IN`).AddArgsGroupStr(candidateIDs).Add(`
			AND pp.type=?`, core.RequirementPagePairType).Add(`
		GROUP BY 1
		HAVING SUM(1)<=SUM(mp.has)
		LIMIT 5`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pageID string
		err := rows.Scan(&pageID)
		if err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		unlockedIDs = append(unlockedIDs, pageID)
		core.AddPageToMap(pageID, returnData.PageMap, core.TitlePlusLoadOptions)
		return nil
	})
	if err != nil {
		return pages.Fail("Error while loading unlocked ids", err)
	}

	// Load pages
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["unlockedIds"] = unlockedIDs
	return pages.Success(returnData)
}

func getHashmapForMasteryInsert(masteryID string, userID string, has bool, wants bool, taughtBy string) database.InsertMap {
	hashmap := make(database.InsertMap)
	hashmap["masteryId"] = masteryID
	hashmap["userId"] = userID
	hashmap["has"] = has
	hashmap["wants"] = wants
	hashmap["createdAt"] = database.Now()
	hashmap["updatedAt"] = database.Now()
	hashmap["taughtBy"] = taughtBy
	return hashmap
}
