// updateMarkHandler.go updates an existing mark.
package site

import (
	"encoding/json"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// updateMarkData contains data given to us in the request.
type updateMarkData struct {
	MarkId string
	Text   string
}

var updateMarkHandler = siteHandler{
	URI:         "/updateMark/",
	HandlerFunc: updateMarkHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// updateMarkHandlerFunc handles requests to create/update a prior like.
func updateMarkHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	var data updateMarkData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Create a new mark
	hashmap := make(database.InsertMap)
	hashmap["id"] = data.MarkId
	hashmap["text"] = data.Text
	statement := db.NewInsertStatement("marks", hashmap, "text")
	_, err = statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't update the mark", err)
	}

	return pages.StatusOK(nil)
}
