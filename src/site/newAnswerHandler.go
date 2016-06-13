// newAnswerHandler.go adds an answer to a question
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
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
	c := params.C
	returnData := core.NewHandlerData(u)
	now := database.Now()

	var data newAnswerData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.QuestionId) || !core.IsIdValid(data.AnswerPageId) {
		return pages.Fail("One of the passed page ids is invalid", nil).Status(http.StatusBadRequest)
	}

	page, err := core.LoadFullEdit(db, data.QuestionId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	}
	if page.Type != core.QuestionPageType {
		return pages.Fail("Adding answer to a non-question page", nil).Status(http.StatusBadRequest)
	}
	if !page.Permissions.Edit.Has {
		return pages.Fail(page.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	var newId int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Add the answer
		hashmap := make(database.InsertMap)
		hashmap["questionId"] = data.QuestionId
		hashmap["answerPageId"] = data.AnswerPageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = now
		statement := db.NewInsertStatement("answers", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert into DB", err)
		}

		newId, err = resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get inserted id", err)
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
		resp, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to changeLogs", err)
		}
		changeLogId, err := resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get changeLog id", err)
		}

		// Insert updates
		var task tasks.NewUpdateTask
		task.UserId = u.Id
		task.GoToPageId = data.AnswerPageId
		task.SubscribedToId = data.QuestionId
		task.UpdateType = core.ChangeLogUpdateType
		task.GroupByPageId = data.QuestionId
		task.ChangeLogId = changeLogId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task: %v", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
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
