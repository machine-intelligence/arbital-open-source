// deleteTagHandler.go handles requests for deleting a tag.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// deleteTagData contains the data we receive in the request.
type deleteTagData struct {
	ParentId int64 `json:",string"`
	ChildId  int64 `json:",string"`
}

// deleteTagHandler handles requests for deleting a tag.
func deleteTagHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	decoder := json.NewDecoder(params.R.Body)
	var data deleteTagData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if data.ParentId == 0 || data.ChildId == 0 {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", err)
	}

	query := db.NewStatement(`
		DELETE FROM pagePairs
		WHERE parentId=? AND childId=? AND type=?`)
	if _, err := query.Exec(data.ParentId, data.ChildId, core.TagPagePairType); err != nil {
		return pages.HandlerErrorFail("Couldn't delete a tag", err)
	}
	return pages.StatusOK(nil)
}
