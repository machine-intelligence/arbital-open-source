// follow.go handles processing a follow.
package tasks

import (
	"fmt"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
	"xelaie/src/go/twitter"
)

// FollowTask is the object that's put into the daemon queue.
type FollowTask struct {
	User      twitter.TwitterUser
	ContestId int64
	CreatedAt string
}

// Check if this task is valid, and we can safely execute it.
func (task *FollowTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *FollowTask) Execute(c sessions.Context) (int, error) {
	c.Debugf("Got a follow by %s", task.User.ScreenName)
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid")
	}

	// TODO: Don't hardcode contestId, but match by the contest owner
	task.ContestId = 2

	hashmap := make(database.InsertMap)
	hashmap["contestId"] = task.ContestId
	hashmap["userId"] = task.User.Id
	hashmap["userName"] = task.User.Name
	hashmap["userScreenName"] = task.User.ScreenName
	hashmap["userFollowers"] = task.User.FollowersCount
	hashmap["createdAt"] = task.CreatedAt

	if err := EnterUserIntoContest(c, task.ContestId, &task.User, hashmap); err != nil {
		return 60, fmt.Errorf("Couldn't register an entree: %v", err)
	}
	return 0, nil
}
