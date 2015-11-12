// newPageJsonHandler.go creates and returns a new page
package site

import (
	"math/rand"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var newPageHandler = siteHandler{
	URI:         "/json/newPage/",
	HandlerFunc: newPageJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// newPageJsonHandler handles the request.
func newPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := newHandlerData(false)

	pageId := rand.Int63()
	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["alias"] = pageId
	hashmap["sortChildrenBy"] = core.LikesChildSortingOption
	hashmap["type"] = core.WikiPageType
	hashmap["maxEdit"] = 1
	hashmap["seeGroupId"] = params.PrivateGroupId
	hashmap["lockedBy"] = u.Id
	hashmap["lockedUntil"] = core.GetPageQuickLockedUntilTime()
	statement := db.NewInsertStatement("pageInfos", hashmap)
	if _, err := statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't update pageInfos", err)
	}

	// Update pages
	hashmap = make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["edit"] = 1
	hashmap["isAutosave"] = true
	hashmap["creatorId"] = u.Id
	hashmap["createdAt"] = database.Now()
	statement = db.NewInsertStatement("pages", hashmap)
	if _, err := statement.Exec(); err != nil {
		return pages.HandlerErrorFail("Couldn't update pages", err)
	}

	core.AddPageIdToMap(pageId, returnData.PageMap)
	return pages.StatusOK(returnData.toJson())
}
