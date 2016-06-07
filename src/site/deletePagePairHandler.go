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
	ParentId string
	ChildId  string
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

	// Get and check input data
	decoder := json.NewDecoder(params.R.Body)
	var data deletePagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if !core.IsIdValid(data.ParentId) || !core.IsIdValid(data.ChildId) {
		return pages.Fail("ParentId and ChildId have to be set", err).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.Fail("Incorrect type", err).Status(http.StatusBadRequest)
	}

	// Load pages.
	parent, err := core.LoadFullEdit(db, data.ParentId, u, nil)
	if err != nil {
		return pages.Fail("Error while loading parent page", err)
	} else if parent == nil {
		parent, err = core.LoadFullEdit(db, data.ParentId, u, &core.LoadEditOptions{LoadNonliveEdit: true})
		if err != nil {
			return pages.Fail("Error while loading parent page (2)", err)
		} else if parent == nil {
			return pages.Fail("Parent page doesn't exist", nil)
		}
	}
	child, err := core.LoadFullEdit(db, data.ChildId, u, nil)
	if err != nil {
		return pages.Fail("Error while loading child page", err)
	} else if child == nil {
		child, err = core.LoadFullEdit(db, data.ChildId, u, &core.LoadEditOptions{LoadNonliveEdit: true})
		if err != nil {
			return pages.Fail("Error while loading child page (2)", err)
		} else if child == nil {
			return pages.Fail("Child page doesn't exist", nil)
		}
	}

	// Check edit permissions
	permissionError, err := core.CanAffectRelationship(c, parent, child, data.Type)
	if err != nil {
		return pages.Fail("Error verifying permissions", err)
	} else if permissionError != "" {
		return pages.Fail(permissionError, nil).Status(http.StatusForbidden)
	}

	// Do it!
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		return deletePagePair(tx, u.Id, data.Type, parent, child)
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	if data.Type == core.ParentPagePairType || data.Type == core.TagPagePairType {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.ChildId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}
	return pages.Success(nil)
}

// deletePagePair deletes the parent-child pagePair of the given type.
func deletePagePair(tx *database.Tx, userId string, pairType string, parent *core.Page, child *core.Page) sessions.Error {
	// Delete the pair
	query := tx.DB.NewStatement(`
			DELETE FROM pagePairs
			WHERE parentId=? AND childId=? AND type=?`).WithTx(tx)
	if _, err := query.Exec(parent.PageId, child.PageId, pairType); err != nil {
		return sessions.NewError("Couldn't delete a page pair", err)
	}

	childIsLive := child.WasPublished && !child.IsDeleted
	parentIsLive := parent.WasPublished && !parent.IsDeleted

	// Update change logs
	var newParentChangeLogId int64

	if childIsLive {
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = parent.PageId
		hashmap["auxPageId"] = child.PageId
		hashmap["userId"] = userId
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.DeleteChildChangeLog,
			core.TagPagePairType:         core.DeleteUsedAsTagChangeLog,
			core.RequirementPagePairType: core.DeleteRequiredByChangeLog,
			core.SubjectPagePairType:     core.DeleteTeacherChangeLog,
		}[pairType]
		statement := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		_, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert new child change log", err)
		}
	}

	if parentIsLive {
		hashmap := make(database.InsertMap)
		hashmap["pageId"] = child.PageId
		hashmap["auxPageId"] = parent.PageId
		hashmap["userId"] = userId
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.DeleteParentChangeLog,
			core.TagPagePairType:         core.DeleteTagChangeLog,
			core.RequirementPagePairType: core.DeleteRequirementChangeLog,
			core.SubjectPagePairType:     core.DeleteSubjectChangeLog,
		}[pairType]
		statement := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
		result, err := statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't insert new child change log", err)
		}
		newParentChangeLogId, err = result.LastInsertId()
		if err != nil {
			return sessions.NewError("Couldn't get changeLogId", err)
		}
	}

	// Send updates for users subscribed to the child.
	if childIsLive && parentIsLive {
		err := tasks.EnqueueNewParentUpdate(tx.DB.C, userId, child.PageId, parent.PageId, newParentChangeLogId)
		if err != nil {
			return sessions.NewError("Couldn't enqueue new parent update", err)
		}
	}

	return nil
}
