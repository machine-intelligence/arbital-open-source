// deleteAnswerHandler.go deletes an answer to a question
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
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
	u := params.U
	db := params.DB

	var data deleteAnswerData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	// Load the existing answer
	answer, err := core.LoadAnswer(db, data.AnswerId)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load the existing answer", err)
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Delete the answer
		statement := database.NewQuery(`
		DELETE FROM answers WHERE id=?`, data.AnswerId).ToStatement(db).WithTx(tx)
		_, err = statement.Exec()
		if err != nil {
			return "Couldn't insert into DB", err
		}

		// Update change logs
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = answer.QuestionId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.AnswerChangeChangeLog
		hashmap["auxPageId"] = answer.AnswerPageId
		hashmap["oldSettingsValue"] = "old"
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't add to changeLogs", err
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	return pages.StatusOK(nil)
}
