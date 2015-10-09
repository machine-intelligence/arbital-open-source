// revertPageHandler.go handles requests for reverting a page. This means marking
// as deleted all autosaves and snapshots which were created by the current user
// after the currently live edit.
package site

import (
	"encoding/json"
	"fmt"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"
)

// revertPageData is the data received from the request.
type revertPageData struct {
	// Page to revert
	PageId int64 `json:",string"`
	// Edit to revert to
	EditNum int
}

// revertPageHandler handles requests for deleting a page.
func revertPageHandler(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data revertPageData
	err := decoder.Decode(&data)
	if err != nil || data.PageId == 0 {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	if !u.IsLoggedIn {
		return pages.HandlerForbiddenFail("Need to be logged in", nil)
	}

	// Load the page
	var page *core.Page
	page, err = loadFullEdit(db, data.PageId, u.Id, &loadEditOptions{loadSpecificEdit: data.EditNum})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	} else if page == nil {
		return pages.HandlerErrorFail("Couldn't find page", nil)
	}

	// Create the data to pass to the edit page handler
	hasVoteStr := ""
	if page.HasVote {
		hasVoteStr = "on"
	}
	parentIds := make([]string, len(page.Parents))
	for n, pair := range page.Parents {
		parentIds[n] = fmt.Sprintf("%d", pair.ParentId)
	}
	editData := &editPageData{
		PageId:         page.PageId,
		PrevEdit:       page.PrevEdit,
		Type:           page.Type,
		Title:          page.Title,
		Clickbait:      page.Clickbait,
		Text:           page.Text,
		MetaText:       page.MetaText,
		HasVoteStr:     hasVoteStr,
		VoteType:       page.VoteType,
		GroupId:        page.GroupId,
		KarmaLock:      page.KarmaLock,
		ParentIds:      strings.Join(parentIds, ","),
		Alias:          page.Alias,
		SortChildrenBy: page.SortChildrenBy,
		AnchorContext:  page.AnchorContext,
		AnchorText:     page.AnchorText,
		AnchorOffset:   page.AnchorOffset,
		RevertToEdit:   data.EditNum,
	}
	return editPageInternalHandler(params, editData)
}
