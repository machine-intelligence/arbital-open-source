// revertPageHandler.go handles requests for reverting a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"
	"net/http"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// revertPageData is the data received from the request.
type revertPageData struct {
	// Page to revert
	PageId string
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

	decoder := json.NewDecoder(params.R.Body)
	var data revertPageData
	err := decoder.Decode(&data)
	if err != nil || !core.IsIdValid(data.PageId) {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}

	// Load the page
	page, err := core.LoadFullEdit(db, data.PageId, u, &core.LoadEditOptions{LoadSpecificEdit: data.EditNum})
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	} else if page == nil {
		return pages.Fail("Couldn't find page", nil)
	}
	if !page.Permissions.Edit.Has {
		return pages.Fail("Can't revert: "+page.Permissions.Edit.Reason, nil).Status(http.StatusBadRequest)
	}

	if page.Type == core.LensPageType {
		// Need to get the actual lens title
		parts := strings.Split(page.Title, ":")
		page.Title = strings.TrimSpace(parts[len(parts)-1])
	}

	// Create the data to pass to the edit page handler
	editData := &editPageData{
		PageId:        page.PageId,
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
