// resolveThreadHandler.go allows a maintainer to mark a comment thread as resolved
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

// resolveThreadData is the data received from the request.
type resolveThreadData struct {
	PageId    string
	Unresolve bool
}

var resolveThreadHandler = siteHandler{
	URI:         "/json/resolveThread/",
	HandlerFunc: resolveThreadHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// resolveThreadHandlerFunc handles requests for deleting a page.
func resolveThreadHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	c := params.C
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data resolveThreadData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.Fail("PageId isn't set", nil).Status(http.StatusBadRequest)
	}

	// Load the page
	page, err := core.LoadFullEdit(db, data.PageId, u, nil)
	if err != nil {
		return pages.Fail("Couldn't load page", err)
	}
	if page.IsDeleted || (page.IsResolved == !data.Unresolve) {
		return pages.Success(nil)
	}
	if page.Type != core.CommentPageType {
		return pages.Fail("Not a comment", nil).Status(http.StatusBadRequest)
	}
	if data.Unresolve {
		// If we are reverting resolve, we need to be able to edit the comment
		if !page.Permissions.Edit.Has {
			return pages.Fail(page.Permissions.Edit.Reason, nil).Status(http.StatusForbidden)
		}
	}

	// Get comment's parents
	commentParentId, commentPrimaryPageId, err := core.GetCommentParents(db, data.PageId)
	if err != nil {
		return pages.Fail("Couldn't load comment's parents", err)
	}
	if commentParentId != data.PageId && commentParentId != "" {
		return pages.Fail("Trying to resolve a reply", nil).Status(http.StatusBadRequest)
	}

	// Only users who have edit access to the comment's primary page can resolve it
	if !data.Unresolve {
		lens, err := core.LoadFullEdit(db, commentPrimaryPageId, u, nil)
		if err != nil {
			return pages.Fail("Couldn't load page", err)
		}
		if !lens.Permissions.Edit.Has {
			return pages.Fail(lens.Permissions.Edit.Reason, nil).Status(http.StatusForbidden)
		}
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Set isResolved in pageInfos
		statement := database.NewQuery(`
			UPDATE pageInfos
			SET isResolved=?`, !data.Unresolve).Add(`
			WHERE pageId=?`, data.PageId).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't set isResolved", err)
		}

		// Notify the comment owner
		var updateTask tasks.NewUpdateTask
		updateTask.UpdateType = core.ResolvedThreadUpdateType
		updateTask.UserId = u.Id
		updateTask.GoToPageId = data.PageId
		updateTask.SubscribedToId = data.PageId
		updateTask.ForceMaintainersOnly = true
		if err := tasks.Enqueue(c, &updateTask, nil); err != nil {
			return sessions.NewError("Couldn't enqueue task", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
