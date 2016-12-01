// approveCommentHandler.go contains the handler for editing pageInfo data.

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

// approveCommentData contains parameters passed in.
type approveCommentData struct {
	CommentID string
}

var approveCommentHandler = siteHandler{
	URI:         "/approveComment/",
	HandlerFunc: approveCommentHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// approveCommentHandlerFunc handles requests to create a new page.
func approveCommentHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U

	// Decode data
	var data approveCommentData
	decoder := json.NewDecoder(params.R.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.CommentID) {
		return pages.Fail("Invalid comment id", nil).Status(http.StatusBadRequest)
	}

	commentParentID, pageID, err := core.GetCommentParents(db, data.CommentID)
	if err != nil {
		return pages.Fail("Couldn't load comment parent page id", err)
	}

	_, canApprove, err := core.CanUserApproveComment(db, u, []string{pageID})
	if err != nil {
		return pages.Fail("Error computing appoving", err)
	} else if !canApprove {
		return pages.Fail("Can't approve comments in this domain", nil).Status(http.StatusBadRequest)
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		hashmap := make(map[string]interface{})
		hashmap["pageId"] = data.CommentID
		hashmap["isApprovedComment"] = true
		statement := tx.DB.NewInsertStatement("pageInfos", hashmap, "isApprovedComment").WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't update pageInfos: %v", err)
		}

		var task tasks.NewUpdateTask
		task.UserID = u.ID
		task.GoToPageID = data.CommentID
		if core.IsIDValid(commentParentID) {
			// This is a new reply
			task.UpdateType = core.ReplyUpdateType
			task.SubscribedToID = commentParentID
		} else {
			// This is a new top level comment
			task.UpdateType = core.TopLevelCommentUpdateType
			task.SubscribedToID = pageID
		}
		if err := tasks.Enqueue(params.C, &task, nil); err != nil {
			return sessions.NewError("Couldn't enqueue a task", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
