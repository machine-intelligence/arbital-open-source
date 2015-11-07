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

var primaryPageHandler = siteHandler{
	URI:         "/json/primaryPage/",
	HandlerFunc: primaryPageJsonHandler,
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
	pageId, ok, err := core.LoadAliasToPageId(db, data.PageAlias)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't convert alias", err)
	}
	if !ok {
		return pages.HandlerErrorFail("Couldn't find page", err)
	}

	// Load data
	returnData := newHandlerData()
	core.AddPageToMap(pageId, returnData.PageMap, core.PrimaryPageLoadOptions)
	err = core.ExecuteLoadPipeline(db, u, returnData.PageMap, returnData.UserMap, returnData.MasteryMap)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	// Computed which pages count as visited. Also update LastVisit if forced
	visitedValues := make([]interface{}, 0)
	for id, p := range returnData.PageMap {
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

	return pages.StatusOK(returnData.toJson())
}
