// updateMasteries.go handles request to add and/or delete masteries

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

// updateMasteries contains the data we get in the request.
type updateMasteries struct {
	// Set mastery id -> given level
	MasteryLevels map[string]int
}

var updateMasteriesHandler = siteHandler{
	URI:         "/updateMasteries/",
	HandlerFunc: updateMasteriesHandlerFunc,
	Options:     pages.PageOptions{},
}

func updateMasteriesHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data updateMasteries
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	return updateMasteriesInternalHandlerFunc(params, &data)
}

func updateMasteriesInternalHandlerFunc(params *pages.HandlerParams, data *updateMasteries) *pages.Result {
	db := params.DB
	u := params.U

	userID := u.GetSomeID()
	if userID == "" {
		return pages.Fail("No user id or session id", nil).Status(http.StatusBadRequest)
	}

	// Translate all mastery aliases to ids
	var masteryAliases []string
	for masteryAlias := range data.MasteryLevels {
		masteryAliases = append(masteryAliases, masteryAlias)
	}
	aliasMap, err := core.LoadAliasToPageIDMap(db, u, masteryAliases)
	if err != nil {
		return pages.Fail("Couldn't translate aliases to ids", err)
	}

	serr := db.Transaction(func(tx *database.Tx) sessions.Error {
		hashmaps := make(database.InsertMaps, 0)
		for masteryAlias, level := range data.MasteryLevels {
			if masteryID, ok := aliasMap[masteryAlias]; ok {
				hashmap := make(database.InsertMap)
				hashmap["masteryId"] = masteryID
				hashmap["userId"] = userID
				hashmap["has"] = true
				hashmap["level"] = level
				hashmap["createdAt"] = database.Now()
				hashmap["updatedAt"] = database.Now()
				hashmaps = append(hashmaps, hashmap)
			}
		}

		if len(hashmaps) > 0 {
			statement := tx.DB.NewMultipleInsertStatement("userMasteryPairs", hashmaps, "has", "level", "updatedAt")
			if _, err := statement.WithTx(tx).Exec(); err != nil {
				return sessions.NewError("Failed to insert masteries", err)
			}
		}
		return nil
	})
	if serr != nil {
		return pages.FailWith(serr)
	}

	return pages.Success(nil)
}
