// deletePagePairHandler.go handles requests for deleting a tag.

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

// deletePagePairData contains the data we receive in the request.
type deletePagePairData struct {
	ParentID string
	ChildID  string
	Type     string
}

var deletePagePairHandler = siteHandler{
	URI:         "/deletePagePair/",
	HandlerFunc: deletePagePairHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// deletePagePairHandlerFunc handles requests for deleting a tag.
func deletePagePairHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U
	handlerData := core.NewHandlerData(u)

	// Get and check input data
	decoder := json.NewDecoder(params.R.Body)
	var data deletePagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIDValid(data.ParentID) || !core.IsIDValid(data.ChildID) {
		return pages.Fail("ParentId and ChildId have to be set", nil).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.Fail("Incorrect type", err).Status(http.StatusBadRequest)
	}

	// Load the page pair
	var pagePair *core.PagePair
	queryPart := database.NewQuery(`WHERE pp.parentId=? AND pp.childId=? AND pp.type=?`, data.ParentID, data.ChildID, data.Type)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pp *core.PagePair) error {
		pagePair = pp
		return nil
	})
	if err != nil {
		return pages.Fail("Failed to load the page pair: %v", err)
	} else if pagePair == nil {
		return pages.Fail("Failed to find the page pair: %v", err)
	}

	// Load pages
	parent, child, err := core.LoadFullEditsForPagePair(db, pagePair, u, handlerData.DomainMap)
	if err != nil {
		return pages.Fail("Error loading pagePair pages", err)
	}

	// Check edit permissions
	permissionError, err := core.CanAffectRelationship(c, parent, child, pagePair.Type)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Do it!
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Delete the pair
		query := tx.DB.NewStatement(`DELETE FROM pagePairs WHERE id=?`).WithTx(tx)
		if _, err := query.Exec(pagePair.ID); err != nil {
			return sessions.NewError("Couldn't delete the page pair", err)
		}

		// If we never published this page pair, there is nothing to do
		if !pagePair.EverPublished {
			return nil
		}

		// Update change logs
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = parent.PageID
		hashmap["auxPageId"] = child.PageID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.DeleteChildChangeLog,
			core.TagPagePairType:         core.DeleteUsedAsTagChangeLog,
			core.RequirementPagePairType: core.DeleteRequiredByChangeLog,
			core.SubjectPagePairType:     core.DeleteTeacherChangeLog,
		}[pagePair.Type]
		statement := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to parent change log", err)
		}
		deletedChildChangeLogID, err := result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get child changeLogId", err)
		}

		hashmap = make(database.InsertMap)
		hashmap["pageId"] = child.PageID
		hashmap["auxPageId"] = parent.PageID
		hashmap["userId"] = u.ID
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.DeleteParentChangeLog,
			core.TagPagePairType:         core.DeleteTagChangeLog,
			core.RequirementPagePairType: core.DeleteRequirementChangeLog,
			core.SubjectPagePairType:     core.DeleteSubjectChangeLog,
		}[pagePair.Type]
		statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		result, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't add to child change log", err)
		}
		deletedParentChangeLogID, err := result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get parent changeLogId", err)
		}

		// Send updates for users subscribed to the parent.
		err = tasks.EnqueuePagePairUpdate(tx.DB.C, pagePair, u.ID, deletedChildChangeLogID, false)
		if err != nil {
			return sessions.NewError("Couldn't enqueue child updates", err)
		}
		// Send updates for users subscribed to the child.
		err = tasks.EnqueuePagePairUpdate(tx.DB.C, pagePair, u.ID, deletedParentChangeLogID, true)
		if err != nil {
			return sessions.NewError("Couldn't enqueue parent updates", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	return pages.Success(nil)
}
