// updateMasteries.go handles request to add and/or delete masteries
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateMasteries contains the data we get in the request.
type updateMasteries struct {
	RemoveMasteries []string
	WantsMasteries  []string
	AddMasteries    []string
}

var updateMasteriesHandler = siteHandler{
	URI:         "/updateMasteries/",
	HandlerFunc: updateMasteriesHanlderFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func updateMasteriesHanlderFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data updateMasteries
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	allMasteries := append(append(data.AddMasteries, data.RemoveMasteries...), data.WantsMasteries...)
	aliasMap, err := core.LoadAliasToPageIdMap(db, allMasteries)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't translate aliases to ids", err)
	}

	// Create values list for creating the query
	// NOTE: order is important: we are deleting, then adding "wants", then adding "haves"
	queryValues := make([]interface{}, 0)
	for _, masteryAlias := range data.RemoveMasteries {
		if masteryId, ok := aliasMap[masteryAlias]; ok {
			queryValues = append(queryValues, masteryId, u.Id, false, false, database.Now(), database.Now())
		}
	}
	for _, masteryAlias := range data.WantsMasteries {
		if masteryId, ok := aliasMap[masteryAlias]; ok {
			queryValues = append(queryValues, masteryId, u.Id, false, true, database.Now(), database.Now())
		}
	}
	for _, masteryAlias := range data.AddMasteries {
		if masteryId, ok := aliasMap[masteryAlias]; ok {
			queryValues = append(queryValues, masteryId, u.Id, true, false, database.Now(), database.Now())
		}
	}

	// Update the database
	if len(queryValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO userMasteryPairs (masteryId,userId,has,wants,createdAt,updatedAt)
			VALUES ` + database.ArgsPlaceholder(len(queryValues), 6) + `
			ON DUPLICATE KEY UPDATE has=VALUES(has),wants=VALUES(wants),updatedAt=VALUES(updatedAt)`)
		if _, err = statement.Exec(queryValues...); err != nil {
			return pages.HandlerErrorFail("Couldn't update masteries", err)
		}
	}

	return pages.StatusOK(nil)
}
