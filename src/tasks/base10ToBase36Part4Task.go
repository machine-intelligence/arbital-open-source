// base10ToBase36Part4Task.go does part 4 of converting all the ids from base 10 to base 36
package tasks

import (
	"zanaduu3/src/database"
)

// Base10ToBase36Part4Task is the object that's put into the daemon queue.
type Base10ToBase36Part4Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part4Task) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part4Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 4 START ====")
	defer c.Debugf("==== PART 4 COMPLETED ====")

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		doOneQuery(db, `ALTER TABLE pages DROP pageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pages DROP creatorIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE changeLogs DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE changeLogs DROP pageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE changeLogs DROP auxPageIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE groupMembers DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE groupMembers DROP groupIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE likes DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE likes DROP pageIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE links DROP parentIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE pageDomainPairs DROP pageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageDomainPairs DROP domainIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE pageInfos DROP pageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP lockedByBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP seeGroupIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP editGroupIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP createdByBase36 ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP aliasBase36 ;`)

		doOneQuery(db, `ALTER TABLE pagePairs DROP parentIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE pagePairs DROP childIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE pageSummaries DROP pageIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE subscriptions DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE subscriptions DROP toIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE updates DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE updates DROP groupByPageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE updates DROP groupByUserIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE updates DROP subscribedToIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE updates DROP goToPageIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE updates DROP byUserIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE userMasteryPairs DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE userMasteryPairs DROP masteryIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE users DROP idBase36 ;`)

		doOneQuery(db, `ALTER TABLE visits DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE visits DROP pageIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE votes DROP userIdBase36 ;`)
		doOneQuery(db, `ALTER TABLE votes DROP pageIdBase36 ;`)

		doOneQuery(db, `ALTER TABLE pages DROP pageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pages DROP creatorIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE changeLogs DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE changeLogs DROP pageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE changeLogs DROP auxPageIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE groupMembers DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE groupMembers DROP groupIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE likes DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE likes DROP pageIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE links DROP parentIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE pageDomainPairs DROP pageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageDomainPairs DROP domainIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE pageInfos DROP pageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP lockedByProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP seeGroupIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP editGroupIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP createdByProcessed ;`)
		doOneQuery(db, `ALTER TABLE pageInfos DROP aliasProcessed ;`)

		doOneQuery(db, `ALTER TABLE pagePairs DROP parentIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE pagePairs DROP childIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE pageSummaries DROP pageIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE subscriptions DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE subscriptions DROP toIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE updates DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE updates DROP groupByPageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE updates DROP groupByUserIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE updates DROP subscribedToIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE updates DROP goToPageIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE updates DROP byUserIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE userMasteryPairs DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE userMasteryPairs DROP masteryIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE users DROP idProcessed ;`)

		doOneQuery(db, `ALTER TABLE visits DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE visits DROP pageIdProcessed ;`)

		doOneQuery(db, `ALTER TABLE votes DROP userIdProcessed ;`)
		doOneQuery(db, `ALTER TABLE votes DROP pageIdProcessed ;`)

		doOneQuery(db, `DROP TABLE pagesandusers;`)
		doOneQuery(db, `DROP TABLE base10tobase36;`)

		return "", nil
	})
	if errMessage != "" {
		return 0, err
	}

	return 0, err
}
