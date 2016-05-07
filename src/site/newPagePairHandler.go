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

	// The following parameters can be ONLY passed from internal callers
	SkipPermissionsCheck bool `json:"-"`
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
	var err error

	// Error checking
	if !core.IsIdValid(data.ParentId) || !core.IsIdValid(data.ChildId) {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", nil)
	}
	if data.ParentId == data.ChildId &&
		data.Type != core.SubjectPagePairType &&
		data.Type != core.RequirementPagePairType {
		return pages.HandlerBadRequestFail("ParentId equals ChildId", nil)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.HandlerBadRequestFail("Incorrect type", err)
	}

	// Load existing connections
	var unusedType string
	exists, _ := database.NewQuery(`
		SELECT type
		FROM pagePairs
		WHERE parentId=?`, data.ParentId).Add(`
			AND childId=?`, data.ChildId).Add(`
			AND type=?`, data.Type).ToStatement(db).QueryRow().Scan(&unusedType)
	if err != nil {
		return pages.HandlerErrorFail("Couldn't load existing pair types", err)
	} else if exists {
		// We already have this type of connection
		return pages.StatusOK(nil)
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

	// Error checking
	// TODO: we can't check this right now because this breaks creating comments in private groups
	/*if child.SeeGroupId != parent.SeeGroupId {
		return pages.HandlerErrorFail("Parent and child need to have the same See Group", nil)
	}*/
	// Check edit permissions
	if !data.SkipPermissionsCheck {
		if data.Type == core.RequirementPagePairType || data.Type == core.TagPagePairType {
			// We don't need permissions for the parent
			delete(pageMap, data.ParentId)
		}
		permissionError, err := core.VerifyPermissionsForMap(db, pageMap, u)
		if err != nil {
			return pages.HandlerForbiddenFail("Error verifying permissions", err)
		} else if permissionError != "" {
			return pages.HandlerForbiddenFail(permissionError, nil)
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

		if child.Edit > 0 && !child.IsDeleted {
			err = addNewChildToChangelog(tx, u.Id, data.Type, child.Type, data.ParentId, parent.Edit, data.ChildId, child.Edit,
				child.IsDeleted)
			if err != nil {
				return "Couldn't add to changelog of parent", err
			}
		}
		if parent.Edit > 0 && !parent.IsDeleted {
			err = addNewParentToChangelog(tx, u.Id, data.Type, child.Type, data.ChildId, child.Edit, data.ParentId, parent.Edit,
				parent.IsDeleted)
			if err != nil {
				return "Couldn't add to changelog of child", err
			}
		}
		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Generate updates for users who are subscribed to the parent/child pages.
	if parent.Type != core.CommentPageType && parent.Alias != "" && child.Alias != "" &&
		!parent.IsDeleted && !child.IsDeleted {

		tasks.EnqueueNewRelationshipUpdates(c, u.Id, data.Type, child.Type, parent.PageId, child.PageId)

		if data.Type == core.ParentPagePairType {
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
func addNewChildToChangelog(tx *database.Tx, userId string, pairType string, childPageType string, pageId string,
	pageEdit int, childId string, childEdit int, childIsDeleted bool) error {

	return addRelationshipToChangelogInternal(tx, userId, pairType, childPageType, pageId, pageEdit, childId, childEdit,
		childIsDeleted, false)
}

// Update the changelogs of the child for a new relationship.
func addNewParentToChangelog(tx *database.Tx, userId string, pairType string, childPageType string, pageId string,
	pageEdit int, parentId string, parentEdit int, parentIsDeleted bool) error {

	return addRelationshipToChangelogInternal(tx, userId, pairType, childPageType, pageId, pageEdit, parentId, parentEdit,
		parentIsDeleted, true)
}

func getChangeLogTypeForPagePair(pairType string, childPageType string, changeLogEntryIsForChild bool) (string, error) {
	switch pairType {
	case core.ParentPagePairType:
		if changeLogEntryIsForChild {
			return core.NewParentChangeLog, nil
		} else if childPageType == core.LensPageType {
			return core.NewLensChangeLog, nil
		} else {
			return core.NewChildChangeLog, nil
		}
	case core.TagPagePairType:
		if changeLogEntryIsForChild {
			return core.NewTagChangeLog, nil
		} else {
			return core.NewUsedAsTagChangeLog, nil
		}
	case core.RequirementPagePairType:
		if changeLogEntryIsForChild {
			return core.NewRequirementChangeLog, nil
		} else {
			return core.NewRequiredByChangeLog, nil
		}
	case core.SubjectPagePairType:
		if changeLogEntryIsForChild {
			return core.NewSubjectChangeLog, nil
		} else {
			return core.NewTeacherChangeLog, nil
		}
	}

	return "", fmt.Errorf("Unexpected pagePair type")
}

func addRelationshipToChangelogInternal(tx *database.Tx, userId string, pairType string, childPageType string, pageId string,
	pageEdit int, auxPageId string, auxPageEdit int, auxPageIsDeleted bool, changeLogEntryIsForChild bool) error {

	// Do not add to the changelog of a public page if its aux page hasn't been published (as this would leak data
	// about a user's unpublished draft) or if it's deleted (editing a deleted page shouldn't affect live pages
	// until the deleted page is published again).
	if auxPageId != pageId && (auxPageEdit <= 0 || auxPageIsDeleted) {
		return nil
	}

	entryType, err := getChangeLogTypeForPagePair(pairType, childPageType, changeLogEntryIsForChild)
	if err != nil {
		return fmt.Errorf("Could not get changelog type for relationship: %v", err)
	}

	hashmap := make(database.InsertMap)
	hashmap["pageId"] = pageId
	hashmap["auxPageId"] = auxPageId
	hashmap["userId"] = userId
	hashmap["edit"] = pageEdit
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = entryType

	_, err = tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx).Exec()
	return err
}
