// commentThreadHandler.go loads and returns all the comments in a comment thread.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// commentThreadData contains parameters passed in to create a page.
type commentThreadData struct {
	CommentId string `json:"pageAlias"`
}

var commentThreadHandler = siteHandler{
	URI:         "/commentThread/",
	HandlerFunc: commentThreadHandlerFunc,
}

// commentThreadHandlerFunc handles requests to create a new page.
func commentThreadHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	returnData := core.NewHandlerData(params.U)

	// Decode data
	var data commentThreadData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.CommentId) {
		return pages.Fail("Need a valid commentId", nil).Status(http.StatusBadRequest)
	}

	_, commentPrimaryPageId, err := core.GetCommentParents(db, data.CommentId)
	if err != nil {
		return pages.Fail("Couldn't load comment's parents", err)
	}

	// Load the comments.
	loadOptions := (&core.PageLoadOptions{
		Parents: true,
	}).Add(core.SubpageLoadOptions)
	core.AddPageToMap(data.CommentId, returnData.PageMap, loadOptions)
	core.AddPageToMap(commentPrimaryPageId, returnData.PageMap, (&core.PageLoadOptions{
		DomainsAndPermissions: true,
	}).Add(core.TitlePlusLoadOptions))
	err = core.ExecuteLoadPipeline(db, returnData)
	if err != nil {
		return pages.Fail("Pipeline error", err)
	}

	return pages.Success(returnData)
}
