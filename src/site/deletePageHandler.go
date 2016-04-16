// deletePageHandler.go handles requests for deleting a page.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/elastic"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// deletePageData is the data received from the request.
type deletePageData struct {
	PageId string
}

var deletePageHandler = siteHandler{
	URI:         "/deletePage/",
	HandlerFunc: deletePageHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// deletePageHandlerFunc handles requests for deleting a page.
func deletePageHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	decoder := json.NewDecoder(params.R.Body)
	var data deletePageData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.PageId) {
		return pages.HandlerBadRequestFail("PageId isn't set", nil)
	}

	// Load the page
	pageMap := make(map[string]*core.Page)
	page := core.AddPageIdToMap(data.PageId, pageMap)
	err = core.LoadPages(db, u, pageMap)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load page", err)
	}
	if page == nil {
		// Looks like there is no need to delete this page.
		return pages.StatusOK(nil)
	}
	if page.Type == core.GroupPageType || page.Type == core.DomainPageType {
		if !u.IsAdmin {
			return pages.HandlerForbiddenFail("Have to be an admin to delete a group/domain", nil)
		}
	}
	if page.Type == core.CommentPageType && u.Id != page.CreatorId {
		if !u.IsAdmin {
			return pages.HandlerForbiddenFail("Have to be an admin to delete someone else's comment", nil)
		}
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Clear the current edit in pages
		statement := tx.DB.NewStatement("UPDATE pages SET isLiveEdit=false WHERE pageId=? AND isLiveEdit").WithTx(tx)
		if _, err = statement.Exec(data.PageId); err != nil {
			return "Couldn't update isLiveEdit for old edits", err
		}

		// Set isDeleted in pageInfos
		statement = tx.DB.NewStatement("UPDATE pageInfos SET isDeleted=true WHERE pageId=?").WithTx(tx)
		if _, err = statement.Exec(data.PageId); err != nil {
			return "Couldn't set isDeleted for deleted page", err
		}

		// Update change log
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = data.PageId
		hashmap["userId"] = u.Id
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = core.DeletePageChangeLog
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't update change logs", err
		}

		// Generate "delete" update for users who are subscribed to this page.
		var task tasks.NewUpdateTask
		task.UserId = u.Id
		task.GoToPageId = data.PageId
		task.SubscribedToId = data.PageId
		task.UpdateType = core.DeletePageUpdateType
		if page.Type == core.CommentPageType {
			_, commentPrimaryPageId, err := core.GetCommentParents(db, data.PageId)
			if err != nil {
				return "Couldn't load comment's parents", err
			}
			task.GroupByPageId = commentPrimaryPageId
		} else {
			task.GroupByPageId = data.PageId
		}
		if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		// Delete it from the elastic index
		if page.WasPublished {
			err = elastic.DeletePageFromIndex(params.C, data.PageId)
			if err != nil {
				return "failed to update index", err
			}
		}
		return "", nil
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(fmt.Sprintf("Transaction failed: %s", errMessage), err)
	}

	// Create a task to propagate the domain change to all children
	var task tasks.PropagateDomainTask
	task.PageId = data.PageId
	if err := tasks.Enqueue(params.C, &task, "propagateDomain"); err != nil {
		return pages.HandlerErrorFail("Couldn't enqueue a task: %v", err)
	}

	return pages.StatusOK(nil)
}
