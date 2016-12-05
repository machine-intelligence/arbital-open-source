// revertPageHandler.go handles requests for reverting a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.

package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// revertPageData is the data received from the request.
type revertPageData struct {
	// Page to revert
	PageID string
	// Edit to revert to
	EditNum int
}

var revertPageHandler = siteHandler{
	URI:         "/revertPage/",
	HandlerFunc: revertPageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// revertPageHandlerFunc handles requests for deleting a page.
func revertPageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	handlerData := core.NewHandlerData(u)

	decoder := json.NewDecoder(params.R.Body)
	var data revertPageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.PageID) {
		return pages.Fail("Missing or invalid page id", nil).Status(http.StatusBadRequest)
	}

	// Load the page
	page, err := core.LoadFullEdit(db, data.PageID, u, handlerData.DomainMap, &core.LoadEditOptions{LoadSpecificEdit: data.EditNum})
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	} else if page == nil {
		return pages.Fail("Couldn't find page", nil)
	}
	if !page.Permissions.Edit.Has {
		return pages.Fail("Can't revert: "+page.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	// Create the data to pass to the edit page handler
	editData := &editPageData{
		PageID:        page.PageID,
		PrevEdit:      page.PrevEdit,
		Title:         page.Title,
		Clickbait:     page.Clickbait,
		Text:          page.Text,
		MetaText:      page.MetaText,
		EditSummary:   page.EditSummary,
		AnchorContext: page.AnchorContext,
		AnchorText:    page.AnchorText,
		AnchorOffset:  page.AnchorOffset,
		RevertToEdit:  data.EditNum,
	}
	return editPageInternalHandler(params, editData)
}
