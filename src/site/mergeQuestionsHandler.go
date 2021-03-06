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
	QuestionID     string
	IntoQuestionID string
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
	if !core.IsIDValid(data.QuestionID) || !core.IsIDValid(data.IntoQuestionID) {
		return pages.Fail("One of the ids is invalid", nil).Status(http.StatusBadRequest)
	}

	// Load the page
	pageMap := make(map[string]*core.Page)
	question := core.AddPageIDToMap(data.QuestionID, pageMap)
	intoQuestion := core.AddPageIDToMap(data.IntoQuestionID, pageMap)
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
			SET questionId=?`, data.IntoQuestionID).Add(`
			WHERE questionId=?`, data.QuestionID).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update answers", err)
		}

		statement = database.NewQuery(`
			UPDATE marks
			SET resolvedPageId=?`, data.IntoQuestionID).Add(`
			WHERE resolvedPageId=?`, data.QuestionID).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update answers", err)
		}

		statement = database.NewQuery(`
			UPDATE pageInfos
			SET mergedInto=?`, data.IntoQuestionID).Add(`
			WHERE pageId=?`, data.QuestionID).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos", err)
		}

		// Delete the question page
		deletePageData := &deletePageData{
			PageID:         data.QuestionID,
			GenerateUpdate: false,
		}
		return deletePageTx(tx, params, deletePageData, question)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Generate "merge" update for users who are subscribed to either of the questions
	var updateTask tasks.NewUpdateTask
	updateTask.UserID = u.ID
	updateTask.GoToPageID = data.IntoQuestionID
	updateTask.SubscribedToID = data.QuestionID
	updateTask.UpdateType = core.QuestionMergedUpdateType
	if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
	updateTask.GoToPageID = data.QuestionID
	updateTask.SubscribedToID = data.IntoQuestionID
	updateTask.UpdateType = core.QuestionMergedReverseUpdateType
	if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}

	return pages.Success(nil)
}
