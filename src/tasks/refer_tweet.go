// refer_email.go handles the task of sending a referral email for the daemon queue.
package tasks

import (
	"fmt"
	"math/rand"
	"time"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
	"xelaie/src/go/twitter"

	"github.com/garyburd/go-oauth/oauth"
)

// ReferEmailTask is the object that's put into the daemon queue.
type ReferTweetTask struct {
	FromId         int64              `json:",string"` // Twitter id
	FromScreenName string             // Twitter @handle of the sender
	Text           string             // tweet text to send
	Creds          *oauth.Credentials // credentials to send the tweet
}

// Check if this task is valid, and we can safely execute it.
func (task *ReferTweetTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *ReferTweetTask) Execute(c sessions.Context) (int, error) {
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid")
	}

	rand.Seed(time.Now().UnixNano())
	referralId := rand.Int63()
	referral := make(database.InsertMap)
	referral["referralId"] = referralId
	referral["fromUserId"] = task.FromId
	referral["fromScreenName"] = task.FromScreenName
	referral["text"] = task.Text
	referral["createdAt"] = database.Now()

	err := database.ExecuteSql(c, database.GetInsertSql("referrals_sent", referral))
	if err != nil {
		return 60, fmt.Errorf("Couldn't add the referral to the db: %v", err)
	}

	text := fmt.Sprintf("%s http://rewards.xelaie.com/?referralId=%d", task.Text, referralId)
	_, err = twitter.UpdateStatusWithCreds(c, task.Creds, text)
	if err != nil {
		return 600, fmt.Errorf("Couldn't update status: %v", err)
	}
	c.Debugf("Tweet sent!")

	return 0, nil
}
