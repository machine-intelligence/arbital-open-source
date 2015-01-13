// update_contest.go handles the "update contest" tasks in the daemon queue.
package tasks

import (
	"fmt"
	"strings"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/twitter"
)

// UpdateContestTask is the object that's put into the daemon queue.
type UpdateContestTask struct {
	Action   string
	Text     string
	Hashtag  string
	StartAt  string
	PayoutId int64 `json:",string"`
}

// Check if this task is valid, and we can safely execute it.
func (task *UpdateContestTask) IsValid() error {
	if !strings.Contains(task.Text, task.Hashtag) {
		return fmt.Errorf("Tweet doesn't contain the hashtag.")
	}
	if task.Text[0] == '@' {
		return fmt.Errorf("Tweet shouldn't start with '@' symbol.")
	}
	if len(task.StartAt) > 0 {
		_, err := time.ParseInLocation(database.TimeLayout, task.StartAt, database.PacificLocation)
		if err != nil {
			return fmt.Errorf("StartAt is in the wrong format: %v", err)
		}
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *UpdateContestTask) Execute(c sessions.Context) (int, error) {
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid")
	}
	if task.Action == "add" {
		if len(task.StartAt) > 0 {
			// Convert time to UTC (validation has been done already)
			startAt, _ := time.ParseInLocation(database.TimeLayout, task.StartAt, database.PacificLocation)
			duration := startAt.UTC().Sub(time.Now()).Seconds()
			if duration > 5.0 {
				// Postpone this task if it's more than 5 seconds in the future, allowing
				// for rounding errors and such.
				return int(duration), nil
			}
		}
		text, err := twitter.UpdateStatus(c, task.Text)
		if err != nil {
			// Retry in 60 seconds
			return 60, fmt.Errorf("Failed to update status: %v", err)
		}
		hashmap := make(map[string]interface{})
		hashmap["ownerId"] = 2804496104
		hashmap["ownerName"] = "Xelaie Rewards"
		hashmap["ownerScreenName"] = "Xelaie"
		hashmap["payoutId"] = task.PayoutId
		hashmap["hashtag"] = task.Hashtag
		hashmap["createdAt"] = database.Now()
		hashmap["text"] = text
		hashmap["isActive"] = true
		sql := database.GetInsertSql("contests", hashmap)
		if err = database.ExecuteSql(c, sql); err != nil {
			// TODO: this sucks, because we already sent out the tweet, so we don't
			// want to retry the whole task. May be this should be spun off a
			// separate queue task.
			return 0, fmt.Errorf("Failed to execute sql command to add a contest: %v", err)
		}
		// TODO: actually, no need to reload them. We just have to add this new one.
		if err = LoadContests(c); err != nil {
			// TODO: yup, this situation kind of sucks too
			return 0, fmt.Errorf("Error reloading contests: %v", err)
		}
	} else if task.Action == "expire" {
	} else {
		return -1, fmt.Errorf("Unknown update contest action: %s", task.Action)
	}
	return 0, nil
}
