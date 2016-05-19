// Handles requests to dismiss updates
package site

import (
	"encoding/json"
	"net/http"

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
	statement := db.NewStatement(`
		UPDATE updates
		SET dismissed=TRUE
		WHERE id=?`)
	if _, err := statement.Exec(data.UpdateId); err != nil {
		return pages.Fail("Couldn't dismiss update", err)
	}

	return pages.Success(nil)
}
