// deletePagePairHandler.go handles requests for deleting a tag.
package site

import (
	"encoding/json"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
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
		MinKarma:     200,
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
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if !core.IsIdValid(data.ParentId) || !core.IsIdValid(data.ChildId) {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", err)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.HandlerBadRequestFail("Incorrect type", err)
	}

	// Load the pages
	pageMap := make(map[string]*core.Page)
	parent := core.AddPageIdToMap(data.ParentId, pageMap)
	child := core.AddPageIdToMap(data.ChildId, pageMap)

	// Load pages.
	err = core.LoadPages(db, u, pageMap)
	if err != nil {
		return pages.HandlerErrorFail("error while loading pages", err)
	}

	// Do it!
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		return deletePagePair(tx, u.Id, parent.PageId, child.PageId, data.Type)
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	if data.Type == core.ParentPagePairType || data.Type == core.TagPagePairType {
		// Create a task to propagate the domain change to all children
		var task tasks.PropagateDomainTask
		task.PageId = data.ChildId
		if err := tasks.Enqueue(c, &task, "propagateDomain"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}
	}
	return pages.StatusOK(nil)
}

// deletePagePair deletes the parent-child pagePair of the given type.
func deletePagePair(tx *database.Tx, userId, parentId, childId string, pairType string) (string, error) {
	// Delete the pair
	query := tx.DB.NewStatement(`
			DELETE FROM pagePairs
			WHERE parentId=? AND childId=? AND type=?`).WithTx(tx)
	if _, err := query.Exec(parentId, childId, pairType); err != nil {
		return "Couldn't delete a page pair", err
	}

	// Update change log
	hashmap := make(database.InsertMap)
	hashmap["pageId"] = parentId
	hashmap["auxPageId"] = childId
	hashmap["userId"] = userId
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = map[string]string{
		core.ParentPagePairType:      core.DeleteChildChangeLog,
		core.TagPagePairType:         core.DeleteUsedAsTagChangeLog,
		core.RequirementPagePairType: core.DeleteRequiredByChangeLog,
		core.SubjectPagePairType:     core.DeleteTeacherChangeLog,
	}[pairType]
	statement := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return "Couldn't insert new child change log", err
	}

	hashmap = make(database.InsertMap)
	hashmap["pageId"] = childId
	hashmap["auxPageId"] = parentId
	hashmap["userId"] = userId
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = map[string]string{
		core.ParentPagePairType:      core.DeleteParentChangeLog,
		core.TagPagePairType:         core.DeleteTagChangeLog,
		core.RequirementPagePairType: core.DeleteRequirementChangeLog,
		core.SubjectPagePairType:     core.DeleteSubjectChangeLog,
	}[pairType]
	statement = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return "Couldn't insert new child change log", err
	}
	return "", nil
}
