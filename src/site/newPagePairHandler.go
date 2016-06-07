// newPagePairHandler.go handles repages for adding a new tag.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
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
	},
}

// newPagePairHandlerFunc handles requests for adding a new tag.
func newPagePairHandlerFunc(params *pages.HandlerParams) *pages.Result {
	decoder := json.NewDecoder(params.R.Body)
	var data newPagePairData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
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
		return pages.Fail("ParentId and ChildId have to be set", nil).Status(http.StatusBadRequest)
	}
	if data.ParentId == data.ChildId &&
		data.Type != core.SubjectPagePairType &&
		data.Type != core.RequirementPagePairType {
		return pages.Fail("ParentId equals ChildId", nil).Status(http.StatusBadRequest)
	}
	data.Type, err = core.CorrectPagePairType(data.Type)
	if err != nil {
		return pages.Fail("Incorrect type", err).Status(http.StatusBadRequest)
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
		return pages.Fail("Couldn't load existing pair types", err)
	} else if exists {
		// We already have this type of connection
		return pages.Success(nil)
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

	// Ids of the changelogs created for this pagePair
	var newChildChangeLogId int64
	var newParentChangeLogId int64

	// Do it!
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Create new page pair
		hashmap := make(database.InsertMap)
		hashmap["parentId"] = data.ParentId
		hashmap["childId"] = data.ChildId
		hashmap["type"] = data.Type
		statement := tx.DB.NewInsertStatement("pagePairs", hashmap).WithTx(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert pagePair", err)
		}

		if child.WasPublished && !child.IsDeleted {
			newChildChangeLogId, err = addNewChildToChangelog(tx, u.Id, data.Type, child.Type, data.ParentId, parent.Edit,
				data.ChildId, child.WasPublished, child.IsDeleted)
			if err != nil {
				return sessions.NewError("Couldn't add to changelog of parent", err)
			}
		}
		if parent.WasPublished && !parent.IsDeleted {
			newParentChangeLogId, err = addNewParentToChangelog(tx, u.Id, data.Type, child.Type, data.ChildId, child.Edit,
				data.ParentId, parent.WasPublished, parent.IsDeleted)
			if err != nil {
				return sessions.NewError("Couldn't add to changelog of child", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	// Generate updates for users who are subscribed to the parent/child pages.
	if parent.Type != core.CommentPageType && child.Type != core.CommentPageType &&
		parent.Alias != "" && child.Alias != "" && !parent.IsDeleted && !child.IsDeleted {

		err := tasks.EnqueueNewParentUpdate(c, u.Id, child.PageId, parent.PageId, newParentChangeLogId)
		if err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		if data.Type == core.ParentPagePairType {
			// Create a task to propagate the domain change to all children
			var task tasks.PropagateDomainTask
			task.PageId = child.PageId
			if err := tasks.Enqueue(c, &task, nil); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.Success(nil)
}

// Update the changelogs of the parent for a new relationship.
func addNewChildToChangelog(tx *database.Tx, userId string, pairType string, childPageType string, pageId string,
	pageEdit int, childId string, childWasPublished bool, childIsDeleted bool) (int64, error) {

	return addRelationshipToChangelogInternal(tx, userId, pairType, childPageType, pageId, pageEdit, childId, childWasPublished,
		childIsDeleted, false)
}

// Update the changelogs of the child for a new relationship.
func addNewParentToChangelog(tx *database.Tx, userId string, pairType string, childPageType string, pageId string,
	pageEdit int, parentId string, parentWasPublished bool, parentIsDeleted bool) (int64, error) {

	return addRelationshipToChangelogInternal(tx, userId, pairType, childPageType, pageId, pageEdit, parentId, parentWasPublished,
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
	pageEdit int, auxPageId string, auxPageWasPublished bool, auxPageIsDeleted bool, changeLogEntryIsForChild bool) (int64, error) {

	// Do not add to the changelog of a public page if its aux page hasn't been published (as this would leak data
	// about a user's unpublished draft) or if it's deleted (editing a deleted page shouldn't affect live pages
	// until the deleted page is published again).
	if auxPageId != pageId && (!auxPageWasPublished || auxPageIsDeleted) {
		return 0, nil
	}
	// Don't add changelog entries for the parents of new comments
	if childPageType == core.CommentPageType && !changeLogEntryIsForChild {
		return 0, nil
	}

	entryType, err := getChangeLogTypeForPagePair(pairType, childPageType, changeLogEntryIsForChild)
	if err != nil {
		return 0, fmt.Errorf("Could not get changelog type for relationship: %v", err)
	}

	hashmap := make(database.InsertMap)
	hashmap["pageId"] = pageId
	hashmap["auxPageId"] = auxPageId
	hashmap["userId"] = userId
	hashmap["edit"] = pageEdit
	hashmap["createdAt"] = database.Now()
	hashmap["type"] = entryType

	result, err := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Could not insert changeLog: %v", err)
	}
	changeLogId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Could not get changeLogId: %v", err)
	}
	return changeLogId, err
}
