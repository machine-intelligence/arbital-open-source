// publishPagePairTask.go will publish a given page pair if it's applicable and
// hasn't been published already.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// PublishPagePairTask is the object that's put into the daemon queue.
type PublishPagePairTask struct {
	UserID     string
	PagePairID string
}

func (task PublishPagePairTask) Tag() string {
	return "publishPagePair"
}

// Check if this task is valid, and we can safely execute it.
func (task PublishPagePairTask) IsValid() error {
	if task.PagePairID == "" {
		return fmt.Errorf("PagePairId needs to be set")
	}
	if task.UserID == "" {
		return fmt.Errorf("UserId needs to be set")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task PublishPagePairTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return -1, err
	}

	c.Infof("==== PUBLISH PAGE PAIR START ====")
	defer c.Infof("==== PUBLISH PAGE PAIR COMPLETED ====")

	// Load the page pair
	var pagePair *core.PagePair
	queryPart := database.NewQuery(`
		WHERE NOT pp.everPublished AND pp.id=?`, task.PagePairID)
	err = core.LoadPagePairs(db, queryPart, func(db *database.DB, pp *core.PagePair) error {
		pagePair = pp
		return nil
	})
	if err != nil {
		return -1, fmt.Errorf("Failed to load the page pair: %v", err)
	} else if pagePair == nil {
		return -1, fmt.Errorf("Failed to find the page pair: %v", err)
	}

	// Load all the involved pages
	pageMap := make(map[string]*core.Page)
	parent := core.AddPageIDToMap(pagePair.ParentID, pageMap)
	child := core.AddPageIDToMap(pagePair.ChildID, pageMap)
	err = core.LoadPages(db, nil, pageMap)
	if err != nil {
		return -1, fmt.Errorf("Failed to load all the pages: %v", err)
	}

	// Only process page pair when both pages are valid
	childIsValid := child.WasPublished && !child.IsDeleted && child.Type != core.CommentPageType
	parentIsValid := parent.WasPublished && !parent.IsDeleted && parent.Type != core.CommentPageType
	if !childIsValid || !parentIsValid {
		return 0, nil
	}

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Mark page pair as published
		hashmap := make(database.InsertMap)
		hashmap["id"] = pagePair.ID
		hashmap["everPublished"] = true
		_, err := tx.DB.NewInsertStatement("pagePairs", hashmap, "everPublished").WithTx(tx).Exec()
		if err != nil {
			return sessions.NewError("Could set everPublished", err)
		}

		// Add change log entries to both pages
		parentChangeLogID, err := addNewRelationshipToParentChangeLog(tx, pagePair, false)
		if err != nil {
			return sessions.NewError("Couldn't add to changelog of parent", err)
		}
		childChangeLogID, err := addNewRelationshipToParentChangeLog(tx, pagePair, true)
		if err != nil {
			return sessions.NewError("Couldn't add to changelog of child", err)
		}

		// Update people subscribed to the parent
		err = EnqueuePagePairUpdate(tx.DB.C, pagePair, task.UserID, parentChangeLogID, false)
		if err != nil {
			tx.DB.C.Errorf("Couldn't enqueue a task: %v", err)
		}
		// Update people subscribed to the child
		err = EnqueuePagePairUpdate(tx.DB.C, pagePair, task.UserID, childChangeLogID, true)
		if err != nil {
			tx.DB.C.Errorf("Couldn't enqueue a task: %v", err)
		}

		if pagePair.Type == core.ParentPagePairType {
			// Create a task to propagate the domain change to all children
			var task PropagateDomainTask
			task.PageID = child.PageID
			if err := Enqueue(c, &task, nil); err != nil {
				tx.DB.C.Errorf("Couldn't enqueue a task: %v", err)
			}
		}
		return nil
	})
	if err2 != nil {
		return -1, sessions.ToError(err2)
	}

	return 0, nil
}

// Give page pair type, return what change log type should be for the parent's change log entry.
// If forChild is set, then for the child's change log entry.
func getChangeLogTypeForPagePair(pairType string, forChild bool) (string, error) {
	switch pairType {
	case core.ParentPagePairType:
		if forChild {
			return core.NewParentChangeLog, nil
		} else {
			return core.NewChildChangeLog, nil
		}
	case core.TagPagePairType:
		if forChild {
			return core.NewTagChangeLog, nil
		} else {
			return core.NewUsedAsTagChangeLog, nil
		}
	case core.RequirementPagePairType:
		if forChild {
			return core.NewRequirementChangeLog, nil
		} else {
			return core.NewRequiredByChangeLog, nil
		}
	case core.SubjectPagePairType:
		if forChild {
			return core.NewSubjectChangeLog, nil
		} else {
			return core.NewTeacherChangeLog, nil
		}
	}

	return "", fmt.Errorf("Unexpected pagePair type")
}

// Update parent's changelog to refrect this new page pair.
// If forChild is passed, will update child's changelog instead.
func addNewRelationshipToParentChangeLog(tx *database.Tx, pagePair *core.PagePair, forChild bool) (int64, error) {
	if (pagePair.Type == core.TagPagePairType || pagePair.Type == core.RequirementPagePairType) && !forChild {
		return 0, nil
	}
	entryType, err := getChangeLogTypeForPagePair(pagePair.Type, forChild)
	if err != nil {
		return 0, fmt.Errorf("Could not get changelog type for relationship: %v", err)
	}

	hashmap := make(database.InsertMap)
	if !forChild {
		hashmap["pageId"] = pagePair.ParentID
		hashmap["auxPageId"] = pagePair.ChildID
	} else {
		hashmap["pageId"] = pagePair.ChildID
		hashmap["auxPageId"] = pagePair.ParentID
	}
	hashmap["type"] = entryType
	hashmap["userId"] = pagePair.CreatorID
	hashmap["createdAt"] = database.Now()
	result, err := tx.DB.NewInsertStatement("changeLogs", hashmap).WithTx(tx).Exec()
	if err != nil {
		return 0, fmt.Errorf("Could not insert changeLog: %v", err)
	}
	changeLogID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Could not get changeLogId: %v", err)
	}
	return changeLogID, err
}

// Add an update to the queue about the pagePair being added to the parent's changeLog.
// If forChild is set, the update is about child's changeLog.
func EnqueuePagePairUpdate(c sessions.Context, pagePair *core.PagePair, userID string, changeLogID int64, forChild bool) error {
	if changeLogID == 0 {
		return nil
	}
	// Don't send updates for pages that are being used as tags or requirements
	if (pagePair.Type == core.TagPagePairType || pagePair.Type == core.RequirementPagePairType) && !forChild {
		return nil
	}

	var task NewUpdateTask
	task.UserID = userID
	task.ChangeLogID = changeLogID
	task.UpdateType = core.ChangeLogUpdateType
	if !forChild {
		task.SubscribedToID = pagePair.ParentID
		task.GoToPageID = pagePair.ChildID
	} else {
		task.SubscribedToID = pagePair.ChildID
		task.GoToPageID = pagePair.ParentID
	}
	return Enqueue(c, &task, nil)
}
