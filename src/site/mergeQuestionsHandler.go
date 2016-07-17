// mergeQuestionsHandler.go merges one question into another.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

// mergeQuestionsData is the data received from the request.
type mergeQuestionsData struct {
	QuestionId     string
	IntoQuestionId string
}

var mergeQuestionsHandler = siteHandler{
	URI:         "/mergeQuestions/",
	HandlerFunc: mergeQuestionsHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

func mergeQuestionsHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	c := params.C
	db := params.DB

	decoder := json.NewDecoder(params.R.Body)
	var data mergeQuestionsData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.QuestionId) || !core.IsIdValid(data.IntoQuestionId) {
		return pages.Fail("One of the ids is invalid", nil).Status(http.StatusBadRequest)
	}

	// Load the page
	pageMap := make(map[string]*core.Page)
	question := core.AddPageIdToMap(data.QuestionId, pageMap)
	intoQuestion := core.AddPageIdToMap(data.IntoQuestionId, pageMap)
	err = core.LoadPages(db, u, pageMap)
	if err != nil {
		return pages.Fail("Couldn't load questions", err)
	}
	if question.Type != core.QuestionPageType {
		return pages.Fail("QuestionId isn't a question", nil).Status(http.StatusBadRequest)
	}
	if intoQuestion.Type != core.QuestionPageType {
		return pages.Fail("IntoQuestionId isn't a question", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		statement := database.NewQuery(`
			UPDATE answers
			SET questionId=?`, data.IntoQuestionId).Add(`
			WHERE questionId=?`, data.QuestionId).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update answers", err)
		}

		statement = database.NewQuery(`
			UPDATE marks
			SET resolvedPageId=?`, data.IntoQuestionId).Add(`
			WHERE resolvedPageId=?`, data.QuestionId).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update answers", err)
		}

		statement = database.NewQuery(`
			UPDATE pageInfos
			SET mergedInto=?`, data.IntoQuestionId).Add(`
			WHERE pageId=?`, data.QuestionId).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Delete the question page
		deletePageData := &deletePageData{
			PageId:         data.QuestionId,
			GenerateUpdate: false,
		}
		return deletePageTx(tx, params, deletePageData, question)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Generate "merge" update for users who are subscribed to either of the questions
	var updateTask tasks.NewUpdateTask
	updateTask.UserId = u.ID
	updateTask.GoToPageId = data.IntoQuestionId
	updateTask.SubscribedToId = data.QuestionId
	updateTask.UpdateType = core.QuestionMergedUpdateType
	if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
	updateTask.GoToPageId = data.QuestionId
	updateTask.SubscribedToId = data.IntoQuestionId
	updateTask.UpdateType = core.QuestionMergedReverseUpdateType
	if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
