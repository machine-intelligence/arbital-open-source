// Handles requests to dismiss updates
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

type dismissUpdateData struct {
	UpdateId int `json:"id"`
}

var dismissUpdateHandler = siteHandler{
	URI:         "/dismissUpdate/",
	HandlerFunc: dismissUpdateHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func dismissUpdateHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	// Decode data
	var data dismissUpdateData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	// Dismiss the update
	hashmap := make(database.InsertMap)
	hashmap["id"] = data.UpdateId
	hashmap["dismissed"] = true
	statement := db.NewInsertStatement("updates", hashmap, "dismissed")
	if _, err := statement.Exec(); err != nil {
		return pages.Fail("Couldn't dismiss the update", err)
	}

	return pages.Success(nil)
}
