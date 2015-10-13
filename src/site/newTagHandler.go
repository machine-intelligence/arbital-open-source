// newTag.go handles repages for adding a new tag.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newTagData contains the data we get in the request.
type newTagData struct {
	ParentId int64 `json:",string"`
	ChildId  int64 `json:",string"`
}

// newTagHandler handles requests for adding a new tag.
func newTagHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data newTagData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.ParentId == 0 || data.ChildId == 0 {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	hashmap := make(database.InsertMap)
	hashmap["parentId"] = data.ParentId
	hashmap["childId"] = data.ChildId
	hashmap["type"] = core.TagPagePairType
	statement := db.NewInsertStatement("pagePairs", hashmap, "parentId")
	_, err = statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't create new tag", err)
	}
	return pages.StatusOK(nil)
}
