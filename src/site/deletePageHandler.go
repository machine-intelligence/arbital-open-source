// deletePageHandler.go handles requests for deleting a page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// deletePageData is the data received from the request.
type deletePageData struct {
	PageId int64 `json:",string"`
}

// deletePageHandler handles requests for deleting a page.
func deletePageHandler(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	decoder := json.NewDecoder(params.R.Body)
	var data deletePageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Have to be logged in", nil)
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, nil)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	}
	if page == nil || !page.WasPublished || page.Type == core.DeletedPageType {
		// Looks like there is no need to delete this page.
		return pages.StatusOK(nil)
	}

	// Create the data to pass to the edit page handler
	hasVoteStr := ""
	if page.HasVote {
		hasVoteStr = "on"
	}
	parentIds := make([]string, len(page.Children))
	for n, pair := range page.Children {
		parentIds[n] = fmt.Sprintf("%d", pair.ChildId)
	}
	editData := &editPageData{
		PageId:         page.PageId,
		PrevEdit:       page.Edit,
		Type:           core.DeletedPageType,
		Title:          "[DELETED]",
		HasVoteStr:     hasVoteStr,
		VoteType:       page.VoteType,
		GroupId:        page.GroupId,
		KarmaLock:      page.KarmaLock,
		Alias:          fmt.Sprintf("%d", page.PageId),
		SortChildrenBy: page.SortChildrenBy,
		DeleteEdit:     true,
	}
	return editPageInternalHandler(params, editData)
}
