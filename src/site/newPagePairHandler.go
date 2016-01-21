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
	ParentId int64 `json:",string"`
	ChildId  int64 `json:",string"`
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
	if data.ParentId <= 0 || data.ChildId <= 0 {
		return pages.HandlerBadRequestFail("ParentId and ChildId have to be set", nil)
	}
	if data.ParentId == data.ChildId && data.Type != core.SubjectPagePairType {
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
	pageMap := make(map[int64]*core.Page)
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
		if data.Type == core.ParentPagePairType && parent.SeeGroupId != 0 && parent.SeeGroupId != child.SeeGroupId {
			return pages.HandlerErrorFail("SeeGroupId has to be the same for parent and child", nil)
		}
		if data.Type == core.RequirementPagePairType && parent.SeeGroupId == 0 && child.SeeGroupId != 0 {
			return pages.HandlerErrorFail("For a public parent, all requirements have to be public", nil)
		}
		if data.Type == core.SubjectPagePairType && parent.SeeGroupId == 0 && child.SeeGroupId != 0 {
			return pages.HandlerErrorFail("For a public parent, all subjects have to be public", nil)
		}
		if child.Type == core.AnswerPageType && parent.Type != core.QuestionPageType {
			return pages.HandlerErrorFail("Answer page can only be a child of a question page", nil)
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
		statement := tx.NewInsertTxStatement("pagePairs", hashmap, "parentId")
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert pagePair", err
		}

		// Update change log
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.ParentId
		hashmap["auxPageId"] = data.ChildId
		hashmap["userId"] = u.Id
		hashmap["edit"] = parent.Edit
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.NewChildChangeLog,
			core.TagPagePairType:         core.NewTagTargetChangeLog,
			core.RequirementPagePairType: core.NewRequiredForChangeLog,
			core.SubjectPagePairType:     core.NewTeachesChangeLog,
		}[data.Type]
		statement = tx.NewInsertTxStatement("changeLogs", hashmap)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert new child change log", err
		}
		hashmap = make(database.InsertMap)
		hashmap["pageId"] = data.ChildId
		hashmap["auxPageId"] = data.ParentId
		hashmap["userId"] = u.Id
		hashmap["edit"] = child.Edit
		hashmap["createdAt"] = database.Now()
		hashmap["type"] = map[string]string{
			core.ParentPagePairType:      core.NewParentChangeLog,
			core.TagPagePairType:         core.NewTagChangeLog,
			core.RequirementPagePairType: core.NewRequirementChangeLog,
			core.SubjectPagePairType:     core.NewSubjectChangeLog,
		}[data.Type]
		statement = tx.NewInsertTxStatement("changeLogs", hashmap)
		if _, err = statement.Exec(); err != nil {
			return "Couldn't insert new child change log", err
		}
		return "", nil
	})
	if err != nil {
		return pages.HandlerErrorFail(errMessage, err)
	}

	// Generate updates for users who are subscribed to the parent pages.
	if parent.Type != core.CommentPageType && parent.Alias != "" && child.Alias != "" {
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
		if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		task = tasks.NewUpdateTask{}
		task.UserId = u.Id
		if data.Type == core.ParentPagePairType {
			task.UpdateType = core.NewChildUpdateType
		} else if data.Type == core.TagPagePairType {
			task.UpdateType = core.NewTaggedByUpdateType
		} else if data.Type == core.RequirementPagePairType {
			task.UpdateType = core.NewRequiredByUpdateType
		} else if data.Type == core.SubjectPagePairType {
			task.UpdateType = core.NewTeachesUpdateType
		}
		task.GroupByPageId = parent.PageId
		task.SubscribedToId = parent.PageId
		task.GoToPageId = child.PageId
		if err := tasks.Enqueue(c, &task, "newUpdate"); err != nil {
			c.Errorf("Couldn't enqueue a task: %v", err)
		}

		if data.Type == core.ParentPagePairType || data.Type == core.TagPagePairType {
			// Create a task to propagate the domain change to all children
			var task tasks.PropagateDomainTask
			task.PageId = child.PageId
			if err := tasks.Enqueue(c, &task, "propagateDomain"); err != nil {
				c.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
	}

	return pages.StatusOK(nil)
}
