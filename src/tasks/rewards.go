// rewards.go loads and manages all the rewards.
package tasks

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
)

const (
	retweetPayoutId  = 1
	followPayoutId   = 2
	referralPayoutId = 10
)

var (
	rewards          []Reward
	rewardsForKids   []Reward
	rewardsByGender  map[string][]Reward
	rewardsByCompany map[string][]Reward
)

type Reward struct {
	Id      int64
	Sex     string
	ForKids bool
	Company sql.NullString
}

// processRewardRow is called for each row when loading rewards from the database.
func processRewardRow(rows *sql.Rows) error {
	var r Reward
	err := rows.Scan(
		&r.Id,
		&r.Sex,
		&r.Company,
		&r.ForKids)
	if err != nil {
		return fmt.Errorf("failed to scan a reward row: %v", err)
	}
	rewards = append(rewards, r)
	return nil
}

// LoadRewards loads all the available rewards from the database.
func LoadRewards(c sessions.Context) error {
	rewards = make([]Reward, 0, 1000)
	sql := "SELECT rewardId, sex, company, forKids FROM rewards WHERE isActive AND type='discount'"
	if err := database.QuerySql(c, sql, processRewardRow); err != nil {
		return fmt.Errorf("Failed to execute sql command to add a reward: %v", err)
	}
	c.Debugf("Loaded %d rewards", len(rewards))

	// Break down rewards in all sorts of ways for various payouts to use.
	rewardsByGender = make(map[string][]Reward)
	rewardsByCompany = make(map[string][]Reward)
	for _, r := range rewards {
		if r.ForKids {
			rewardsForKids = append(rewardsForKids, r)
			continue
		}
		rewardsByGender[r.Sex] = append(rewardsByGender[r.Sex], r)
		if r.Company.Valid {
			rewardsByCompany[r.Company.String] = append(rewardsByCompany[r.Company.String], r)
		}
	}
	// TODO: print this info more dynamically / automatically
	c.Debugf("======F: %v", len(rewardsByGender["F"]))
	c.Debugf("======M: %v", len(rewardsByGender["M"]))
	c.Debugf("======O: %v", len(rewardsByGender["O"]))
	c.Debugf("======K: %v", len(rewardsForKids))
	c.Debugf("======C1: %v", len(rewardsByCompany["Dick's Sporting Goods"]))
	c.Debugf("======C2: %v", len(rewardsByCompany["Aeropostale"]))
	c.Debugf("======C3: %v", len(rewardsByCompany["Bambeco"]))
	return nil
}

// randInt returns a random integer [min, max]
func randInt(min int, max int) int {
	return int(rand.Int31n(int32(max-min+1))) + min
}

// generateRewards returns an array of random rewards.
func GenerateRewards(payoutId int64, followers int) (rewardIds []int64, copper int) {
	// TODO: pull reward data for this contest from our database. For now it's all hardcoded.
	// Generate rewards.
	rewardIds = make([]int64, 0, 100)
	// Make sure everyone has a chance to win any prize, but the chances go up
	// with the number of followers.
	const MIL = 1000000
	rand.Seed(time.Now().UnixNano())
	r := randInt(1, MIL)

	// Giftcards
	cardOdds1, cardOdds2, cardOdds3 := 0, 0, 0
	if payoutId == followPayoutId || payoutId == referralPayoutId {
		cardOdds1 = 1 + followers/40
		cardOdds2 = 10 + followers/5
		cardOdds3 = 100 + followers*4
	} else {
		cardOdds1 = 1 + followers/80
		cardOdds2 = 10 + followers/10
		cardOdds3 = 100 + followers*1
	}
	if payoutId == referralPayoutId { // Referral
		cardOdds3 += cardOdds2 + cardOdds1
		cardOdds2 = 0
		cardOdds1 = 0
	}
	if r > MIL-cardOdds1 {
		// $100 card
		rewardIds = append(rewardIds, 1)
	} else if r > MIL-cardOdds1-cardOdds2 {
		// $25 card
		rewardIds = append(rewardIds, 2)
	} else if r > MIL-cardOdds1-cardOdds2-cardOdds3 {
		// $5 card
		rewardIds = append(rewardIds, 3)
	}

	// Select which rewards we are using.
	curRewards := rewards
	if payoutId == 3 {
		curRewards = rewardsByGender["F"]
	} else if payoutId == 4 {
		curRewards = rewardsByGender["M"]
	} else if payoutId == 5 {
		curRewards = rewardsByGender["O"]
	} else if payoutId == 6 {
		curRewards = rewardsForKids
	} else if payoutId == 7 {
		curRewards = rewardsByCompany["Dick's Sporting Goods"]
	} else if payoutId == 8 {
		curRewards = rewardsByCompany["Aeropostale"]
	} else if payoutId == 9 {
		curRewards = rewardsByCompany["Bambeco"]
	}

	// Rewards
	rewardCount := randInt(3, 4)
	for n := 0; n < rewardCount; n++ {
		rewardId := curRewards[randInt(0, len(curRewards)-1)].Id
		rewardIds = append(rewardIds, rewardId)
	}

	// For now, we are not giving out any copper, since the user doesn't really
	// see that it was added as part of a reward for a specific action.
	copper = 0

	return
}
