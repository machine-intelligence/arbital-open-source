// primaryPageJsonHandler.go contains the handler for returning JSON with data
// to display a primary page.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// primaryPageJsonData contains parameters passed in via the request.
type primaryPageJsonData struct {
	PageAlias       string
	ForcedLastVisit string
}

// primaryPageJsonHandler handles the request.
func primaryPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data primaryPageJsonData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	// Get actual page id
	aliasToIdMap, err := core.LoadAliasToPageIdMap(db, []string{data.PageAlias})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	pageId, ok := aliasToIdMap[data.PageAlias]
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	// Load data
	userMap := make(map[int64]*core.User)
	pageMap := make(map[int64]*core.Page)
	masteryMap := make(map[int64]*core.Mastery)

	core.AddPageToMap(pageId, pageMap, core.PrimaryPageLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Computed which pages count as visited. Also update LastVisit if forced
	visitedValues := make([]interface{}, 0)
	for id, p := range pageMap {
		if p.Text != "" {
			visitedValues = append(visitedValues, u.Id, id, database.Now())
		}
		if data.ForcedLastVisit != "" && p.LastVisit > data.ForcedLastVisit {
			p.LastVisit = data.ForcedLastVisit
		}
	}

	// Add a visit to pages for which we loaded text.
	if len(visitedValues) > 0 {
		statement := db.NewStatement(`
			INSERT INTO visits (userId, pageId, createdAt)
			VALUES ` + database.ArgsPlaceholder(len(visitedValues), 3))
		if _, err = statement.Exec(visitedValues...); err != nil {
			return pages.HandlerErrorFail("Couldn't update visits", err)
		}
	}

	returnData := createReturnData(pageMap).AddUsers(userMap).AddMasteries(masteryMap)
	return pages.StatusOK(returnData)
}
