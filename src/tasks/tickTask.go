// tick.go updates all the computed values in our database.
package tasks

import (
	"database/sql"
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	tickPeriod = 3600
)

// TickTask is the object that's put into the daemon queue.
type TickTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *TickTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *TickTask) Execute(c sessions.Context) (delay int, err error) {
	delay = tickPeriod
	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== TICK START ====")
	// Compute all priors.
	err = database.QuerySql(c, `SELECT id FROM questions`, processQuestion)
	return
}

func processQuestion(c sessions.Context, rows *sql.Rows) error {
	var questionId int64
	if err := rows.Scan(&questionId); err != nil {
		return fmt.Errorf("failed to scan for questionId: %v", err)
	}
	c.Debugf("==TICK: processing question: %d", questionId)

	// Compute new prior values. Load all most recent prior votes for this question from all users.
	var sumVotes float64
	var votes int64
	query := fmt.Sprintf(`
		SELECT value FROM (
			SELECT userId,value
			FROM priorVotes
			WHERE questionId=%d
			ORDER BY createdAt DESC
		) AS pv
		GROUP BY userId`, questionId)
	err := database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var value float64
		err := rows.Scan(&value)
		if err != nil {
			return fmt.Errorf("failed to scan for prior vote's value: %v", err)
		}
		sumVotes += value
		votes += 1
		return nil
	})
	if err != nil {
		return err
	}
	newValue := float32(sumVotes / float64(votes))

	// Find prior support objects for this question.
	priorMap := make(map[int]int64) // answerIndex -> priorId
	query = fmt.Sprintf(`SELECT id,answerIndex FROM support WHERE questionId=%d AND prior IS NOT NULL`, questionId)
	err = database.QuerySql(c, query, func(c sessions.Context, rows *sql.Rows) error {
		var id int64
		var answerIndex int
		err := rows.Scan(&id, &answerIndex)
		if err != nil {
			return fmt.Errorf("failed to scan for support: %v", err)
		}
		priorMap[answerIndex] = id
		return nil
	})
	if err != nil {
		return err
	}

	// Update priors
	for a := 1; a <= 2; a++ {
		hashmap := make(map[string]interface{})
		hashmap["id"] = priorMap[a]
		if a == 1 {
			hashmap["prior"] = newValue
		} else {
			hashmap["prior"] = 100.0 - newValue
		}
		sql := database.GetInsertSql("support", hashmap, "prior")
		if _, err = database.ExecuteSql(c, sql); err != nil {
			return err
		}
	}

	c.Debugf("==== TICK COMPLETED SUCCESSFULLY ====")
	return nil
}
