// newAnswerHandler.go adds an answer to a question
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
)

// newAnswerData contains data given to us in the request.
type newAnswerData struct {
	QuestionId   string
	AnswerPageId string
}

var newAnswerHandler = siteHandler{
	URI:         "/newAnswer/",
	HandlerFunc: newAnswerHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newAnswerHandlerFunc handles requests to create/update a like.
func newAnswerHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)
	now := database.Now()

	var data newAnswerData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.QuestionId) || !core.IsIdValid(data.AnswerPageId) {
		return pages.HandlerBadRequestFail("One of the passed page ids is invalid", nil)
	}

	page, err := core.LoadFullEdit(db, data.QuestionId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	}
	if page.Type != core.QuestionPageType {
		return pages.HandlerBadRequestFail("Adding answer to a non-question page", nil)
	}
	if !page.Permissions.Edit.Has {
		return pages.HandlerBadRequestFail(page.Permissions.Edit.Reason, nil)
	}

	var newId int64
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Add the answer
		hashmap := make(database.InsertMap)
		hashmap["questionId"] = data.QuestionId
		hashmap["answerPageId"] = data.AnswerPageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = now
		statement := db.NewInsertStatement("answers", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return "Couldn't insert into DB", err
		}

		newId, err = resp.LastInsertId()
		if err != nil {
			return "Couldn't get inserted id", err
		}

		// Update change logs
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.QuestionId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.AnswerChangeChangeLog
		hashmap["auxPageId"] = data.AnswerPageId
		hashmap["newSettingsValue"] = "new"
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't add to changeLogs", err
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.Fail(errMessage, err)
	}

	// Load pages.
	core.AddPageToMap(data.AnswerPageId, returnData.PageMap, core.AnswerLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["newAnswer"], err = core.LoadAnswer(db, fmt.Sprintf("%d", newId))
	if err != nil {
		return pages.Fail("Couldn't load the new answer", err)
	}
	return pages.Success(returnData)
}
