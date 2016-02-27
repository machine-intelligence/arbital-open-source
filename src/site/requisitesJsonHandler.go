// requisitesJsonHandler.go returns all the requisites the user knows
package site

import (
	"encoding/json"
	"net/url"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var requisitesHandler = siteHandler{
	URI:         "/json/requisites/",
	HandlerFunc: requisitesJsonHandler,
	Options: pages.PageOptions{
		LoadUpdateCount: true,
	},
}

func requisitesJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	returnData := core.NewHandlerData(params.U, true)

	// Options to load the pages with
	pageOptions := (&core.PageLoadOptions{}).Add(core.TitlePlusLoadOptions)

	// Load all masteries
	rows := db.NewStatement(`
		SELECT masteryId
		FROM userMasteryPairs
		WHERE userId=?`).Query(u.Id)
	_, err := core.LoadPageIds(rows, returnData.PageMap, pageOptions)
	if err != nil {
		return pages.HandlerErrorFail("Error while loading masteries", err)
	}

	// Load masteryMap from the cookie, if present
	cookie, err := params.R.Cookie("masteryMap")
	if err == nil {
		masteryMap := make(map[string]*core.Mastery)
		jsonStr, _ := url.QueryUnescape(cookie.Value)
		err = json.Unmarshal([]byte(jsonStr), &masteryMap)
		if err != nil {
			params.C.Warningf("Couldn't unmarshal masteryMap cookie: %v", err)
		}

		for masteryId, _ := range masteryMap {
			core.AddPageToMap(masteryId, returnData.PageMap, pageOptions)
		}
	}

	// Load pages.
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	return pages.StatusOK(returnData.ToJson())
}
