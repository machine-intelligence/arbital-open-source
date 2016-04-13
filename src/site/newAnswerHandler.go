// newAnswerHandler.go adds an answer to a question
package site

import (
	"encoding/json"

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
	returnData := core.NewHandlerData(params.U, false)
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

	hashmap := make(map[string]interface{})
	hashmap["questionId"] = data.QuestionId
	hashmap["answerPageId"] = data.AnswerPageId
	hashmap["userId"] = u.Id
	hashmap["createdAt"] = now
	statement := db.NewInsertStatement("answers", hashmap)
	resp, err := statement.Exec()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't insert into DB", err)
	}

	newId, err := resp.LastInsertId()
	if err != nil {
		return pages.HandlerErrorFail("Couldn't get inserted id", err)
	}

	// Load pages.
	core.AddPageToMap(data.AnswerPageId, returnData.PageMap, core.AnswerLoadOptions)
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.HandlerErrorFail("Pipeline error", err)
	}

	returnData.ResultMap["newAnswer"] = core.Answer{
		Id:           newId,
		AnswerPageId: data.AnswerPageId,
		UserId:       u.Id,
		CreatedAt:    now,
	}
	return pages.StatusOK(returnData.ToJson())
}
