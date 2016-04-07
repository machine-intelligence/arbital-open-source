// newMarkHandler.go creates a new mark.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newMarkData contains data given to us in the request.
type newMarkData struct {
	PageId        string
	Edit          int
	AnchorContext string
	AnchorText    string
	AnchorOffset  int
}

var newMarkHandler = siteHandler{
	URI:         "/newMark/",
	HandlerFunc: newMarkHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newMarkHandlerFunc handles requests to create/update a prior like.
func newMarkHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(params.U, true)

	var data newMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("Invalid page id", nil)
	}
	if data.AnchorContext == "" {
		return pages.HandlerBadRequestFail("No anchor context is set", nil)
	}

	hashmap := make(map[string]interface{})
	hashmap["pageId"] = data.PageId
	hashmap["edit"] = data.Edit
	hashmap["anchorContext"] = data.AnchorContext
	hashmap["anchorText"] = data.AnchorText
	hashmap["anchorOffset"] = data.AnchorOffset
	hashmap["creatorId"] = u.Id
	hashmap["createdAt"] = database.Now()
	statement := db.NewInsertStatement("marks", hashmap)
	resp, err := statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't insert an new mark", err)
	}

	lastInsertId, err := resp.LastInsertId()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't get inserted id", err)
	}
	returnData.ResultMap["markId"] = lastInsertId

	return pages.StatusOK(returnData.ToJson())
}
