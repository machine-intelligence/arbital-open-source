// tweet.go handles processing a tweet.
package tasks

import (
	"fmt"
	"strings"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
	"xelaie/src/go/twitter"
)

// TweetTask is the object that's put into the daemon queue.
type TweetTask struct {
	User            twitter.TwitterUser
	TweetId         int64
	Text, CreatedAt string
}

// Check if this task is valid, and we can safely execute it.
func (task *TweetTask) IsValid() error {
	if len(task.Text) <= 0 {
		return fmt.Errorf("Tweet text is empty.")
	}
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *TweetTask) Execute(c sessions.Context) (int, error) {
	c.Debugf("Processing tweet task for @%s", task.User.ScreenName)
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid: %v", err)
	}

	// Simplify the tweet text for broader matching.
	simpleText := simplifyContestText(task.Text)
	if simpleText[0] == '@' {
		// TODO: handle .@ correctly
		// This is an @-mention, so we ignore it completely.
		return 0, nil
	}

	hashmap := make(database.InsertMap)
	hashmap["tweetId"] = task.TweetId
	hashmap["userId"] = task.User.Id
	hashmap["userName"] = task.User.Name
	hashmap["userScreenName"] = task.User.ScreenName
	hashmap["userFollowers"] = task.User.FollowersCount
	hashmap["createdAt"] = task.CreatedAt

	// Find the context with the matching text.
	var contest *Contest
	for _, curContest := range contests {
		if !curContest.Hashtag.Valid {
			// Skip contests with no hashtag, since they are for new followers.
			continue
		}
		// Twitter (and sometimes users) adds stuff to the message. So we just
		// check if the message we need is present.
		if strings.Index(simpleText, curContest.Text.String) >= 0 {
			c.Debugf("Found matching contest: %d", curContest.Id)
			contest = curContest
			break
		}
	}
	if contest == nil {
		c.Debugf("Text doesn't match: %s", simpleText)
		// If the tweet mentions us, let's log it just in case. This way we can see
		// if users are doing something funny.
		if strings.Contains(simpleText, "@xelaie") {
			hashmap["text"] = task.Text
			return 0, database.ExecuteSql(c, database.GetInsertSql("rejected_entrees", hashmap))
		}
		return 0, nil
	}
	hashmap["contestId"] = contest.Id
	c.Debugf("Found a matching contest: %d", contest.Id)
	if err := EnterUserIntoContest(c, contest.Id, &task.User, hashmap); err != nil {
		return -1, fmt.Errorf("Couldn't add an entree: %v", err)
	}
	return 0, nil
}

// simplifyContestText removes extraneous characters from the contest text for broader matching.
// If you check rejected_entrees table, you'll see that for some reasons our
// users RT our tweets with various slight modifications. This is a persistent
// enough issues that we should try to address on our side.
func simplifyContestText(text string) string {
	// TODO: handle ".@" tweets correctly
	clean := func(r rune) rune {
		if strings.ContainsRune(" ~`!$%^&*()-_+=[{]}\\|;:\"',<.>/?", r) {
			return -1
		}
		return r
	}
	return strings.ToLower(strings.Map(clean, text))
}
