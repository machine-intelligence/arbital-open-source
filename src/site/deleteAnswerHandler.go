// deleteAnswerHandler.go deletes an answer to a question
package site

import (
	"encoding/json"

	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// deleteAnswerData contains data given to us in the request.
type deleteAnswerData struct {
	AnswerId string
}

var deleteAnswerHandler = siteHandler{
	URI:         "/deleteAnswer/",
	HandlerFunc: deleteAnswerHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// deleteAnswerHandlerFunc handles requests to create/update a like.
func deleteAnswerHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	var data deleteAnswerData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	statement := database.NewQuery(`
		DELETE FROM answers WHERE id=?`, data.AnswerId).ToStatement(db)
	_, err = statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't insert into DB", err)
	}

	return pages.StatusOK(nil)
}
