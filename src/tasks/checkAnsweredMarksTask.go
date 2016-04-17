// checkAnsweredMarksTask.go checks if the given marks have an answer.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
)

const (
	checkAnsweredMarkPeriod = 1 * 60 * 60 // 1 hour
)

// CheckAnsweredMarksTask is the object that's put into the daemon queue.
type CheckAnsweredMarksTask struct {
}

func (task CheckAnsweredMarksTask) Tag() string {
	return "checkAnsweredMarks"
}

// Check if this task is valid, and we can safely execute it.
func (task CheckAnsweredMarksTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task CheckAnsweredMarksTask) Execute(db *database.DB) (delay int, err error) {
	delay = checkAnsweredMarkPeriod
	c := db.C

	if err := task.IsValid(); err != nil {
		return -1, err
	}

	c.Debugf("==== CHECK ANSWERED MARK START ====")
	defer c.Debugf("==== CHECK ANSWERED MARK COMPLETED ====")

	markIds := make([]string, 0)
	markMap := make(map[string]bool)
	hashmaps := make(database.InsertMaps, 0)

	// Load all unanswered marks that now have an answer
	rows := database.NewQuery(`
		SELECT m.id,m.pageId,m.creatorId
		FROM marks AS m
		JOIN answers AS a
		ON (a.questionId=m.resolvedPageId)
		WHERE NOT m.answered`).ToStatement(db).Query()
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var markId, markPageId, userId string
		if err := rows.Scan(&markId, &markPageId, &userId); err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		_, exists := markMap[markId]
		if !exists {
			markMap[markId] = true
			markIds = append(markIds, markId)

			// Add an update
			hashmap := make(database.InsertMap)
			hashmap["userId"] = userId
			hashmap["type"] = core.AnsweredMarkUpdateType
			hashmap["groupByPageId"] = markPageId
			hashmap["subscribedToId"] = markPageId
			hashmap["goToPageId"] = markPageId
			hashmap["markId"] = markId
			hashmap["createdAt"] = database.Now()
			hashmap["unseen"] = true
			hashmaps = append(hashmaps, hashmap)
		}

		return nil
	})
	if err != nil {
		return -1, fmt.Errorf("Failed to load marks: %v", err)
	}

	if len(markIds) <= 0 {
		return
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		// Update the answered marks
		statement := database.NewQuery(`
			UPDATE marks
			SET answered=true
			WHERE id IN`).AddArgsGroupStr(markIds).ToTxStatement(tx)
		if _, err = statement.Exec(); err != nil {
			return "Failed to load update marks: %v", err
		}

		// Insert all the updates
		statement = tx.DB.NewMultipleInsertStatement("updates", hashmaps)
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return "Couldn't insert into updates", err
		}
		return "", nil
	})
	if err != nil {
		return -1, fmt.Errorf("%s: %v", errMessage, err)
	}

	return
}
