// newPageJsonHandler.go creates and returns a new page
package site

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

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

// newPageJsonData contains parameters passed in via the request.
type newPageJsonData struct {
	Type      string
	ParentIds []string
}

// newPageJsonHandler handles the request.
func newPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data newPageJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode request", err)
	}

	data.Type, err = core.CorrectPageType(data.Type)
	if err != nil {
		data.Type = core.WikiPageType
	}

	pageId := rand.Int63()
	// Update pageInfos
	hashmap := make(map[string]interface{})
	hashmap["pageId"] = pageId
	hashmap["alias"] = pageId
	hashmap["sortChildrenBy"] = core.LikesChildSortingOption
	hashmap["type"] = data.Type
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

	// Add parents
	for _, parentIdStr := range data.ParentIds {
		parentId, err := strconv.ParseInt(parentIdStr, 10, 64)
		if err != nil {
			return pages.HandlerErrorFail(fmt.Sprintf("Invalid parent id: %s", parentId), nil)
		}
		handlerData := newPagePairData{
			ParentId: parentId,
			ChildId:  pageId,
			Type:     core.ParentPagePairType,
		}
		result := newPagePairHandlerInternal(params, &handlerData)
		if result.Message != "" {
			return result
		}
	}

	editData := &editJsonData{PageAlias: fmt.Sprintf("%d", pageId)}
	return editJsonInternalHandler(params, editData)
}
