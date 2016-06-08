// indexJsonHandler.go serves the index page data.
package site

import (
	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

const (
	indexHandlerUri = "/json/index/"
)

type featuredDomain struct {
	DomainId string   `json:"domainId"`
	ChildIds []string `json:"childIds"`
}

var indexHandler = siteHandler{
	URI:         indexHandlerUri,
	HandlerFunc: indexJsonHandler,
	Options:     pages.PageOptions{},
}

func indexJsonHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB
	returnData := core.NewHandlerData(u).SetResetEverything()

	// Load pages.
	core.AddPageIdToMap("3d", returnData.PageMap)
	core.AddPageIdToMap("1sl", returnData.PageMap)
	core.AddPageIdToMap("1sm", returnData.PageMap)
	core.AddPageIdToMap("3hs", returnData.PageMap)
	err := core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	// Add a visit to pages for which we loaded text.
	visitorId := u.GetSomeId()
	if visitorId != "" {
		hashmap := make(database.InsertMap)
		hashmap["userId"] = visitorId
		hashmap["sessionId"] = u.SessionId
		hashmap["ipAddress"] = params.R.RemoteAddr
		hashmap["pageId"] = indexHandlerUri
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("visits", hashmap)
		_, err = statement.Exec()
		if err != nil {
			return pages.Fail("Couldn't insert /index/ visit", err)
		}
	}

	return pages.Success(returnData)
}
