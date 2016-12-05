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
	QuestionID   string
	AnswerPageID string
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

	var data newAnswerData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.QuestionID) || !core.IsIDValid(data.AnswerPageID) {
		return pages.Fail("One of the passed page ids is invalid", nil).Status(http.StatusBadRequest)
	}

	page, err := core.LoadFullEdit(db, data.QuestionID, u, returnData.DomainMap, nil)
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	}
	if page.Type != core.QuestionPageType {
		return pages.Fail("Adding answer to a non-question page", nil).Status(http.StatusBadRequest)
	}
	if !page.Permissions.Edit.Has {
		return pages.Fail(page.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	var newID int64
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Add the answer
		hashmap := make(database.InsertMap)
		hashmap["questionId"] = data.QuestionID
		hashmap["answerPageId"] = data.AnswerPageID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		statement := db.NewInsertStatement("answers", hashmap).WithTx(tx)
		resp, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert into DB", err)
		}

		newID, err = resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get inserted id", err)
		}

		// Update change logs
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.QuestionID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.AnswerChangeChangeLog
		hashmap["auxPageId"] = data.AnswerPageID
		hashmap["newSettingsValue"] = "new"
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		resp, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to changeLogs", err)
		}
		changeLogID, err := resp.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get changeLog id", err)
		}

		// Insert updates
		var task tasks.NewUpdateTask
		task.UserID = u.ID
		task.GoToPageID = data.AnswerPageID
		task.SubscribedToID = data.QuestionID
		task.UpdateType = core.ChangeLogUpdateType
		task.ChangeLogID = changeLogID
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task: %v", err)
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Load pages.
	core.AddPageToMap(data.AnswerPageID, returnData.PageMap, core.AnswerLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	returnData.ResultMap["newAnswer"], err = core.LoadAnswer(db, fmt.Sprintf("%d", newID))
	if err != nil {
		return pages.Fail("Couldn't load the new answer", err)
	}
	return pages.Success(returnData)
}
