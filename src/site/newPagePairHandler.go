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

		// Update change log
		addRelationshipToChangelog(tx, u.Id, data.Type, data.ChildId, data.ParentId, child.Edit, parent.Edit,
			child.IsDeleted, parent.IsDeleted)
		if err != nil {
			return "Couldn't insert new change log", err
		}
		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Generate updates for users who are subscribed to the parent pages.
	if parent.Type != core.CommentPageType && parent.Alias != "" && child.Alias != "" &&
		!parent.IsDeleted && !child.IsDeleted {

		var task tasks.NewUpdateTask
		if data.Type == core.ParentPagePairType {
			task.UpdateType = core.NewParentUpdateType
		} else if data.Type == core.TagPagePairType {
			task.UpdateType = core.NewTagUpdateType
		} else if data.Type == core.RequirementPagePairType {
			task.UpdateType = core.NewRequirementUpdateType
		} else if data.Type == core.SubjectPagePairType {
			task.UpdateType = core.NewSubjectUpdateType
		}
		task.UserId = u.Id
		task.GroupByPageId = child.PageId
		task.SubscribedToId = child.PageId
		task.GoToPageId = parent.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		task = tasks.NewUpdateTask{}
		task.UserId = u.Id
		if data.Type == core.ParentPagePairType {
			task.UpdateType = core.NewChildUpdateType
		} else if data.Type == core.TagPagePairType {
			task.UpdateType = core.NewUsedAsTagUpdateType
		} else if data.Type == core.RequirementPagePairType {
			task.UpdateType = core.NewRequiredByUpdateType
		} else if data.Type == core.SubjectPagePairType {
			task.UpdateType = core.NewTeacherUpdateType
		}
		task.GroupByPageId = parent.PageId
		task.SubscribedToId = parent.PageId
		task.GoToPageId = child.PageId
		if err := tasks.Enqueue(c, &task, nil); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

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

func addRelationshipToChangelog(tx *database.Tx, userId string, pairType string, childId string, parentId string,
	childEdit int, parentEdit int, childIsDeleted bool, parentIsDeleted bool) error {

	// Don't add to the changelog of the parent if the child is unpublished
	if childId == parentId || childEdit > 0 && !childIsDeleted {
		err := addRelationshipToChangelogInternal(tx, userId, pairType, parentId, childId, parentEdit, childEdit,
			core.NewChildChangeLog, core.NewUsedAsTagChangeLog, core.NewRequiredByChangeLog, core.NewTeacherChangeLog)
		if err != nil {
			return err
		}
	}
	// Don't add to the changelog of the child if the parent is unpublished
	if parentId == childId || parentEdit > 0 && !parentIsDeleted {
		err := addRelationshipToChangelogInternal(tx, userId, pairType, childId, parentId, childEdit, parentEdit,
			core.NewParentChangeLog, core.NewTagChangeLog, core.NewRequirementChangeLog, core.NewSubjectChangeLog)
		if err != nil {
			return err
		}
	}
	return nil
}

func addRelationshipToChangelogInternal(tx *database.Tx, userId string, pairType string, pageId string, auxPageId string,
	pageEdit int, auxPageEdit int, parentPPT string, tagPPT string, requirementPPT string, subjectPPT string) error {
	// Do not add to the changelog of a public page if its aux page hasn't been published (as this would leak data
	// about a user's unpublished draft).
	if auxPageEdit <= 0 {
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
