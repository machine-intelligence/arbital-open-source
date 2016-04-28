// newPagePairHandler.go handles repages for adding a new tag.
package site

import (
	"encoding/json"
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/tasks"
)

// newPagePairData contains the data we get in the request.
type newPagePairData struct {
	ParentId string
	ChildId  string
	Type     string
}

var newPagePairHandler = siteHandler{
	URI:         "/newPagePair/",
	HandlerFunc: newPagePairHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
		MinKarma:     200,
	},
}

// newPagePairHandlerFunc handles requests for adding a new tag.
func newPagePairHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data newPagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}

	return newPagePairHandlerInternal(params, &data)
}

func newPagePairHandlerInternal(params *pages.HandlerParams, data *newPagePairData) *pages.Result {
	c := params.C
	db := params.DB
	u := params.U

	// Error checking
	if !core.IsIdValid(data.ParentId) || !core.IsIdValid(data.ChildId) {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", nil)
	}
	if data.ParentId == data.ChildId &&
		data.Type != core.SubjectPagePairType &&
		data.Type != core.RequirementPagePairType {
		return pages.HandlerBadRequestFail("ParentId equals ChildId", nil)
	}
	var err error
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.HandlerBadRequestFail("Incorrect type", err)
	}

	// Load existing connections
	existingTypes := make(map[string]bool)
	rows := database.NewQuery(`
		SELECT type
		FROM pagePairs
		WHERE parentId=? and childId=?
		`, data.ParentId, data.ChildId).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var pairType string
		err := rows.Scan(&pairType)
		if err != nil {
			return fmt.Errorf("Failed to scan pagePairs: %v", err)
		}
		existingTypes[pairType] = true
		return nil
	})
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load existing pair types", err)
	}
	if _, ok := existingTypes[data.Type]; ok {
		// We already have this type of connection
		return pages.StatusOK(nil)
	}

	// Load user groups
	if err := core.LoadUserGroupIds(db, u); err != nil {
		return pages.HandlerForbiddenFail("Couldn't load user groups", err)
	}

	// Load the pages
	pageMap := make(map[string]*core.Page)
	parent := core.AddPageIdToMap(data.ParentId, pageMap)
	child := core.AddPageIdToMap(data.ChildId, pageMap)

	// Load pages.
	err = core.LoadPages(db, u, pageMap)
	if err != nil {
		return pages.HandlerErrorFail("Error while loading pages", err)
	}

	// More error checking
	// TODO: handle cases where either parent or child (or both) are unpublished
	if parent.Alias != "" && child.Alias != "" {
		if data.Type == core.ParentPagePairType && core.IsIdValid(parent.SeeGroupId) && parent.SeeGroupId != child.SeeGroupId {
			return pages.HandlerErrorFail("SeeGroupId has to be the same for parent and child", nil)
		}
		if data.Type == core.RequirementPagePairType && !core.IsIdValid(parent.SeeGroupId) && child.SeeGroupId != "" {
			return pages.HandlerErrorFail("For a public parent, all requirements have to be public", nil)
		}
		if data.Type == core.SubjectPagePairType && !core.IsIdValid(parent.SeeGroupId) && child.SeeGroupId != "" {
			return pages.HandlerErrorFail("For a public parent, all subjects have to be public", nil)
		}
		if child.SeeGroupId != parent.SeeGroupId {
			return pages.HandlerErrorFail("Parent and child need to have the same See Group", nil)
		}
	}

	// Do it!
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Create new page pair
		hashmap := make(database.InsertMap)
		hashmap["parentId"] = data.ParentId
		hashmap["childId"] = data.ChildId
		hashmap["type"] = data.Type
		statement := tx.DB.NewInsertStatement("pagePairs", hashmap, "parentId").WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert pagePair", err
		}

		err = addNewChildToChangelog(tx, u.Id, data.Type, data.ParentId, parent.Edit, data.ChildId, child.Edit, child.IsDeleted)
		if err != nil {
			return "Couldn't add to changelog of parent", err
		}
		err = addNewParentToChangelog(tx, u.Id, data.Type, data.ChildId, child.Edit, data.ParentId, parent.Edit, parent.IsDeleted)
		if err != nil {
			return "Couldn't add to changelog of child", err
		}
		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Generate updates for users who are subscribed to the parent/child pages.
	if parent.Type != core.CommentPageType && parent.Alias != "" && child.Alias != "" &&
		!parent.IsDeleted && !child.IsDeleted {

		tasks.EnqueueNewRelationshipUpdates(c, u.Id, data.Type, parent.PageId, child.PageId)

		if data.Type == core.ParentPagePairType || data.Type == core.TagPagePairType {
			// Create a task to propagate the domain change to all children
			var task tasks.PropagateDomainTask
			task.PageId = child.PageId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.StatusOK(nil)
}

// Update the changelogs of the parent for a new relationship.
func addNewChildToChangelog(tx *database.Tx, userId string, pairType string, pageId string, pageEdit int,
	childId string, childEdit int, childIsDeleted bool) error {
	return addRelationshipToChangelogInternal(tx, userId, pairType, pageId, pageEdit, childId, childEdit, childIsDeleted,
		core.NewChildChangeLog, core.NewUsedAsTagChangeLog, core.NewRequiredByChangeLog, core.NewTeacherChangeLog)
}

// Update the changelogs of the child for a new relationship.
func addNewParentToChangelog(tx *database.Tx, userId string, pairType string, pageId string, pageEdit int,
	parentId string, parentEdit int, parentIsDeleted bool) error {
	return addRelationshipToChangelogInternal(tx, userId, pairType, pageId, pageEdit, parentId, parentEdit, parentIsDeleted,
		core.NewParentChangeLog, core.NewTagChangeLog, core.NewRequirementChangeLog, core.NewSubjectChangeLog)
}

func addRelationshipToChangelogInternal(tx *database.Tx, userId string, pairType string, pageId string, pageEdit int,
	auxPageId string, auxPageEdit int, auxPageIsDeleted bool, parentPPT string, tagPPT string, requirementPPT string,
	subjectPPT string) error {
	// Do not add to the changelog of a public page if its aux page hasn't been published (as this would leak data
	// about a user's unpublished draft) or if it's deleted (editing a deleted page shouldn't affect live pages
	// until the deleted page is published again).
	if auxPageId != pageId && (auxPageEdit <= 0 || auxPageIsDeleted) {
		return nil
	}

	hashmap := make(database.InsertMap)
	hashmap["pageId"] = pageId
	hashmap["auxPageId"] = auxPageId
	hashmap["userId"] = userId
	hashmap["edit"] = pageEdit
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = map[string]string{
		core.ParentPagePairType:      parentPPT,
		core.TagPagePairType:         tagPPT,
		core.RequirementPagePairType: requirementPPT,
		core.SubjectPagePairType:     subjectPPT,
	}[pairType]

	_, err := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx).Exec()
	return err
}
