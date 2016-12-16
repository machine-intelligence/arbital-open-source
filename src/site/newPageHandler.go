// newPageHandler.go creates and returns a new page

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

var newPageHandler = siteHandler{
	URI:         "/json/newPage/",
	HandlerFunc: newPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// newPageData contains parameters passed in via the request.
type newPageData struct {
	Type             string
	ParentIDs        []string
	IsEditorComment  bool
	Alias            string
	SubmitToDomainID string
	// If creating a new comment, this is the id of the primary page
	CommentPrimaryPageID string
}

// newPageHandlerFunc handles the request.
func newPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	// Decode data
	var data newPageData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPageType(data.Type)
	if err != nil {
		data.Type = core.WikiPageType
	}
	return newPageInternalHandler(params, &data)
}

func newPageInternalHandler(params *pages.HandlerParams, data *newPageData) *pages.Result {
	u := params.U

	pageID, err := core.CreateNewPage(params.DB, params.U, &core.CreateNewPageOptions{
		Alias:                data.Alias,
		Type:                 data.Type,
		SeeDomainID:          params.PrivateDomain.ID,
		EditDomainID:         u.MyDomainID(),
		SubmitToDomainID:     data.SubmitToDomainID,
		IsEditorComment:      data.IsEditorComment,
		ParentIDs:            data.ParentIDs,
		CommentPrimaryPageID: data.CommentPrimaryPageID,
	})
	if err != nil {
		return pages.Fail("Couldn't create new page", err)
	}

	editData := &editJSONData{
		PageAlias: pageID,
	}
	return editJSONInternalHandler(params, editData)
}
