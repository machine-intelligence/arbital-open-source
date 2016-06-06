// newPageJsonHandler.go creates and returns a new page
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
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
	Alias           string
}

// newPageJsonHandler handles the request.
func newPageJsonHandler(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data newPageJsonData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPageType(data.Type)
	if err != nil {
		data.Type = core.WikiPageType
	}
	return newPageJsonInternalHandler(params, &data)
}

func newPageJsonInternalHandler(params *pages.HandlerParams, data *newPageJsonData) *pages.Result {
	db := params.DB
	u := params.U

	if data.Alias != "" && !core.IsAliasValid(data.Alias) {
		return pages.Fail("Invalid alias", nil).Status(http.StatusBadRequest)
	}

	// Error checking
	if data.IsEditorComment && data.Type != core.CommentPageType {
		return pages.Fail("Can't set isEditorComment for non-comment pages", nil).Status(http.StatusBadRequest)
	}

	pageId := ""
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		var err error
		pageId, err = core.GetNextAvailableId(tx)
		if err != nil {
			return sessions.NewError("Couldn't get next available Id", err)
		}
		if data.Alias == "" {
			data.Alias = pageId
		}

		// Update pageInfos
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = pageId
		hashmap["alias"] = data.Alias
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
			return sessions.NewError("Couldn't update pageInfos", err)
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
			return sessions.NewError("Couldn't update pages", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Add parents
	for _, parentIdStr := range data.ParentIds {
		handlerData := newPagePairData{
			ParentId: parentIdStr,
			ChildId:  pageId,
			Type:     core.ParentPagePairType,
		}
		result := newPagePairHandlerInternal(params, &handlerData)
		if result.Err != nil {
			return result
		}
	}

	editData := &editJsonData{
		PageAlias: pageId,
	}
	return editJsonInternalHandler(params, editData)
}
