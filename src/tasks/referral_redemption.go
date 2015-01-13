// refer_redemption.go redeems made referrals.
package tasks

import (
	"database/sql"
	"fmt"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
	"xelaie/src/go/twitter"
)

const (
	// Id of the contest we use to mark the user rewards for referrals
	referralContestId = 74
)

type ReferralRedemptionTask struct {
	User           twitter.TwitterUser // current twitter user
	ReferralId     int64               // referral id (if <0, then none)
	FromUserId     int64               // twitter user id of the person who made the referral
	FromScreenName string              // twitter handle of the user who made the referral
	Error          string              // set to a non empty string if there was an error with the referral
}

// Check if this task is valid, and we can safely execute it.
func (task *ReferralRedemptionTask) IsValid() error {
	// TODO: check that either
	// 1) ReferralId is set
	// 2) FromUserId is set
	return nil
}

// Check if this task is valid, and we can safely execute it.
// For comments on return value see tasks.QueueTask
func (task *ReferralRedemptionTask) PreliminaryCheck(c sessions.Context) (int, error) {
	// Check if this user already was referred.
	var fromScreenName string
	sql := fmt.Sprintf("SELECT fromScreenName FROM referrals_redeemed WHERE toUserId=%d", task.User.Id)
	exists, err := database.QueryRowSql(c, sql, &fromScreenName)
	if err != nil {
		task.Error = "Temporary error with referrals. Please try again later."
		return 60, fmt.Errorf("Error while checking if a user was referred: %v", err)
	} else if exists {
		task.Error = fmt.Sprintf("You have already been referred by @%s.", fromScreenName)
		return 0, fmt.Errorf("This user was already referred.")
	}
	return 0, nil
}

// FillByReferralId fllls in the referral info for the task
// when the user was referred by a referral with a generated id.
// For comments on return value see tasks.QueueTask
func (task *ReferralRedemptionTask) FillByReferralId(c sessions.Context) (int, error) {
	c.Debugf("User was referred with id: %d", task.ReferralId)
	// Get information about this referral.
	sql := fmt.Sprintf("SELECT fromUserId, fromScreenName FROM referrals_sent WHERE referralId=%d", task.ReferralId)
	exists, err := database.QueryRowSql(c, sql, &task.FromUserId, &task.FromScreenName)
	if err != nil {
		task.Error = "Internal error. Please try again later."
		return 60, fmt.Errorf("Couldn't check for an existing referral: %v", err)
	} else if !exists {
		task.Error = "You are using an uknown referral number."
		return 0, fmt.Errorf("Unknown referral id: %d", task.ReferralId)
	}
	return 0, nil
}

// FillByReferralUserId fllls in the referral info for the task
// when the user was referred by another user using their id in the link.
// For comments on return value see tasks.QueueTask
func (task *ReferralRedemptionTask) FillByReferralUserId(c sessions.Context) (int, error) {
	c.Debugf("User was referred by user with id: %d", task.FromUserId)
	// Get information about this referral.
	var fromScreenName sql.NullString
	sql := fmt.Sprintf("SELECT userScreenName FROM users WHERE userId=%d", task.FromUserId)
	exists, err := database.QueryRowSql(c, sql, &fromScreenName)
	if err != nil {
		task.Error = "Internal error. Please try again later."
		return 60, fmt.Errorf("Couldn't check for a user: %v", err)
	} else if !exists {
		task.Error = "You are using an uknown referral user id number."
		return 0, fmt.Errorf("Unknown referral user id: %d", task.FromUserId)
	}
	if fromScreenName.Valid {
		task.FromScreenName = fromScreenName.String
	} else {
		c.Warningf("Don't have user's screen name; user id: %d", task.FromUserId)
		task.FromScreenName = fmt.Sprintf("id=%d", task.FromUserId)
	}
	return 0, nil
}

// ProcessReferral does last bit of checkign that the referral is correct and
// then processes it, giving rewards to both parties.
// For comments on return value see tasks.QueueTask
func (task *ReferralRedemptionTask) ProcessReferral(c sessions.Context) (int, error) {
	if task.FromUserId == task.User.Id {
		c.Warningf("User trying to refer themselves.")
		task.Error = "Can't refer yourself. :)"
		return 0, fmt.Errorf("User trying to refer themselves.")
	}

	// Legit referral. Add it to referrals_redeemed.
	referral := make(database.InsertMap)
	if task.ReferralId > 0 {
		referral["referralId"] = task.ReferralId
	}
	referral["toUserId"] = task.User.Id
	referral["fromUserId"] = task.FromUserId
	referral["fromScreenName"] = task.FromScreenName
	referral["createdAt"] = database.Now()
	err := database.ExecuteSql(c, database.GetInsertSql("referrals_redeemed", referral))
	if err != nil {
		task.Error = "Temporary error with the referral. Please try again later."
		return 60, fmt.Errorf("Error updating referrals: %v", err)
	} else {
		// Generate rewards for the referred user
		if err := RewardUser(c, referralContestId, &task.User); err != nil {
			task.Error = "Error generating rewards for your referral. Please contact support."
			return 0, fmt.Errorf("Error generating rewards for a referral: %v", err)
		}
		// Generate rewards for the referral sender
		// For generating rewards, use the number of followers the new user has
		senderUser := twitter.TwitterUser{Id: task.FromUserId, FollowersCount: task.User.FollowersCount}
		if err := RewardUser(c, referralContestId, &senderUser); err != nil {
			return 0, fmt.Errorf("Error generating rewards for a referral: %v", err)
		}
	}
	return 0, nil
}

// Execute this task.
// For comments on return value see tasks.QueueTask
func (task *ReferralRedemptionTask) Execute(c sessions.Context) (int, error) {
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid")
	}

	timeout, err := task.PreliminaryCheck(c)
	if err != nil {
		return timeout, err
	}

	if task.ReferralId > 0 {
		// This is a referral made via email or a tweet
		timeout, err = task.FillByReferralId(c)
	} else {
		// This is a referral made by copying a link with user's id
		timeout, err = task.FillByReferralUserId(c)
	}
	if err != nil {
		return timeout, err
	}

	return task.ProcessReferral(c)
}
