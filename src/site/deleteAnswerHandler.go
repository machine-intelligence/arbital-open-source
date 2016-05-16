// deleteAnswerHandler.go deletes an answer to a question
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
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
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Load the existing answer
	answer, err := core.LoadAnswer(db, data.AnswerId)
	if err != nil {
		return pages.Fail("Couldn't load the existing answer", err)
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Delete the answer
		statement := database.NewQuery(`
			DELETE FROM answers WHERE id=?`, data.AnswerId).ToStatement(db).WithTx(tx)
		_, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert into DB", err)
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
			return sessions.NewError("Couldn't add to changeLogs", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
