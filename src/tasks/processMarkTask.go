// processMarkTask.go processes all the given mark and generates the updates.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

// ProcessMarkTask is the object that's put into the daemon queue.
type ProcessMarkTask struct {
	Id int64
}

func (task *ProcessMarkTask) Tag() string {
	return "processMark"
}

// Check if this task is valid, and we can safely execute it.
func (task *ProcessMarkTask) IsValid() error {
	if task.Id <= 0 {
		return fmt.Errorf("Invalid id: %d", task.Id)
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *ProcessMarkTask) Execute(db *database.DB) (int, error) {
	c := db.C

	if err := task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== MARK UPDATE START ====")
	defer c.Debugf("==== MARK UPDATE COMPLETED ====")

	// Read the mark.
	var pageId, creatorId string
	_, err := db.NewStatement(`
		SELECT pageId,creatorId
		FROM marks
		WHERE id=?`).QueryRow(task.Id).Scan(&pageId, &creatorId)
	if err != nil {
		return -1, err
	}

	var updateTask NewUpdateTask
	updateTask.UserId = creatorId
	updateTask.GoToPageId = pageId
	updateTask.SubscribedToId = pageId
	updateTask.UpdateType = core.NewMarkUpdateType
	updateTask.GroupByPageId = pageId
	updateTask.MarkId = task.Id
	updateTask.EditorsOnly = true
	if err := Enqueue(c, &updateTask, nil); err != nil {
		return -1, fmt.Errorf("Couldn't enqueue an updateTask: %v", err)
	}

	return 0, nil
}
