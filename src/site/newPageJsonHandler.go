// newPageJsonHandler.go creates and returns a new page
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

var newPageHandler = siteHandler{
	URI:         "/json/newPage/",
	HandlerFunc: newPageJsonHandler,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newPageJsonData contains parameters passed in via the request.
type newPageJsonData struct {
	Type            string
	ParentIds       []string
	IsEditorComment bool
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

	// Error checking
	if data.IsEditorComment && data.Type != core.CommentPageType {
		return pages.HandlerBadRequestFail("Can't set isEditorComment for non-comment pages", err)
	}

	pageId := ""
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		pageId, err = core.GetNextAvailableId(tx)
		if err != nil {
			return "", fmt.Errorf("Couldn't get next available Id", err)
		}

		// Update pageInfos
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["alias"] = pageId
		hashmap["sortChildrenBy"] = core.LikesChildSortingOption
		hashmap["type"] = data.Type
		hashmap["maxEdit"] = 1
		hashmap["createdBy"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["seeGroupId"] = params.PrivateGroupId
		hashmap["lockedBy"] = u.Id
		hashmap["lockedUntil"] = core.GetPageQuickLockedUntilTime()
		if data.IsEditorComment {
			hashmap["isEditorComment"] = true
			hashmap["isEditorCommentIntention"] = true
		}
		statement := db.NewInsertStatement("pageInfos", hashmap)
		if _, err := statement.Exec(); err != nil {
			return "", fmt.Errorf("Couldn't update pageInfos", err)
		}

		// Update pages
		hashmap = make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["edit"] = 1
		hashmap["prevEdit"] = 0
		hashmap["isAutosave"] = true
		hashmap["creatorId"] = u.Id
		hashmap["createdAt"] = database.Now()
		statement = db.NewInsertStatement("pages", hashmap)
		if _, err := statement.Exec(); err != nil {
			return "", fmt.Errorf("Couldn't update pages", err)
		}
		return "", err
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Add parents
	for _, parentIdStr := range data.ParentIds {
		handlerData := newPagePairData{
			ParentId: parentIdStr,
			ChildId:  pageId,
			Type:     core.ParentPagePairType,
		}
		result := newPagePairHandlerInternal(params, &handlerData)
		if result.Message != "" {
			return result
		}
	}

	editData := &editJsonData{PageAlias: pageId}
	return editJsonInternalHandler(params, editData)
}
