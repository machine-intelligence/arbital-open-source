// checkAnsweredMarksTask.go checks if the given marks have an answer.
package tasks

import (
	"fmt"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
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

	c.Infof("==== CHECK ANSWERED MARK START ====")
	defer c.Infof("==== CHECK ANSWERED MARK COMPLETED ====")

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
		var markID, markPageID, userID string
		if err := rows.Scan(&markID, &markPageID, &userID); err != nil {
			return fmt.Errorf("Failed to scan: %v", err)
		}
		_, exists := markMap[markID]
		if !exists {
			markMap[markID] = true
			markIds = append(markIds, markID)

			// Add an update
			hashmap := make(database.InsertMap)
			hashmap["userId"] = userID
			hashmap["type"] = core.AnsweredMarkUpdateType
			hashmap["groupByPageId"] = markPageID
			hashmap["subscribedToId"] = markPageID
			hashmap["goToPageId"] = markPageID
			hashmap["markId"] = markID
			hashmap["createdAt"] = database.Now()
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

	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {
		// Update the answered marks
		statement := database.NewQuery(`
			UPDATE marks
			SET answered=true,answeredAt=NOW()
			WHERE id IN`).AddArgsGroupStr(markIds).ToTxStatement(tx)
		if _, err = statement.Exec(); err != nil {
			return sessions.NewError("Failed to load update marks", err)
		}

		// Insert all the updates
		statement = tx.DB.NewMultipleInsertStatement("updates", hashmaps)
		if _, err := statement.WithTx(tx).Exec(); err != nil {
			return sessions.NewError("Couldn't insert into updates", err)
		}
		return nil
	})
	if err2 != nil {
		return -1, sessions.ToError(err2)
	}

	return
}
