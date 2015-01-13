// helpers.go contains helper functions for daemon.
package tasks

import (
	"fmt"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/twitter"
)

// EnterUserIntoContest creates an entree for the given user, also generating rewards.
func EnterUserIntoContest(c sessions.Context,
	contestId int64,
	user *twitter.TwitterUser,
	hashmap database.InsertMap) error {

	// TODO: put everything into a transaction
	err := UpdateUsersTable(c, user)
	if err != nil {
		return fmt.Errorf("Couldn't add an entree: %v", err)
	}

	// Check that we haven't added this user already.
	sql := fmt.Sprintf("SELECT 1 FROM entrees WHERE contestId=%d AND userId=%d", contestId, user.Id)
	var exists bool
	exists, err = database.QueryRowSql(c, sql, &exists)
	if err != nil {
		return fmt.Errorf("Error occured while checking if a user participated in a contest: %v", err)
	}
	if exists {
		c.Debugf("This user already participated in this contest.")
		return nil
	}

	// Add this entree.
	err = database.ExecuteSql(c, database.GetInsertSql("entrees", hashmap))
	if err != nil {
		return fmt.Errorf("Couldn't add an entree: %v", err)
	}
	// Reward user.
	return RewardUser(c, contestId, user)
}

// UpdateUsersTable should be called when we detect that a user explicitly
// interacted with our system. We'll update the users table, adding this uer
// if necessary.
func UpdateUsersTable(c sessions.Context, user *twitter.TwitterUser) error {
	dbUser := make(database.InsertMap)
	dbUser["userId"] = user.Id
	dbUser["userName"] = user.Name
	dbUser["userScreenName"] = user.ScreenName
	dbUser["userFollowers"] = user.FollowersCount
	dbUser["createdAt"] = database.Now()
	dbUser["lastActive"] = database.Now()
	dbUser["lastWebsiteVisit"] = database.Now()
	updateVars := []string{"userName", "userScreenName", "userFollowers", "lastActive", "lastWebsiteVisit"}
	sql := database.GetInsertSql("users", dbUser, updateVars...)
	return database.ExecuteSql(c, sql)
}

// rewardUser generates and gives the appropriate rewards to the user.
func RewardUser(c sessions.Context, contestId int64, user *twitter.TwitterUser) error {
	contest := contests[contestId]
	rewardIds, copper := GenerateRewards(contest.PayoutId, user.FollowersCount)

	// Create all the sql commands for adding rewards.
	sqlCommands := make([]string, 0, len(rewardIds)+1)
	dbReward := make(database.InsertMap)
	dbReward["userId"] = user.Id
	dbReward["contestId"] = contest.Id
	for _, rewardId := range rewardIds {
		dbReward["rewardId"] = rewardId
		sqlCommands = append(sqlCommands, database.GetInsertSql("userRewards", dbReward))
	}

	if copper > 0 {
		copperSql := fmt.Sprintf("UPDATE users SET copper = copper + %d where userId = %d", copper, user.Id)
		sqlCommands = append(sqlCommands, copperSql)
	}
	return database.ExecuteSql(c, sqlCommands...)
}
