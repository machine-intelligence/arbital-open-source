// base10ToBase36Part3Task.go does part 3 of converting all the ids from base 10 to base 36
package tasks

import (
	//"fmt"
	//"regexp"
	//"strings"

	"zanaduu3/src/database"
	//"zanaduu3/src/user"
)

// Base10ToBase36Part3Task is the object that's put into the daemon queue.
type Base10ToBase36Part3Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part3Task) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part3Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 3 START ====")
	defer c.Debugf("==== PART 3 COMPLETED ====")

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		doOneQuery(db, `UPDATE pages SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pages SET creatorId = CONCAT("zzz", creatorId) WHERE 1;`)

		doOneQuery(db, `UPDATE changeLogs SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET auxPageId = CONCAT("zzz", auxPageId) WHERE 1;`)

		doOneQuery(db, `UPDATE groupMembers SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE groupMembers SET groupId = CONCAT("zzz", groupId) WHERE 1;`)

		doOneQuery(db, `UPDATE likes SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE likes SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE links SET parentId = CONCAT("zzz", parentId) WHERE 1;`)

		doOneQuery(db, `UPDATE pageDomainPairs SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageDomainPairs SET domainId = CONCAT("zzz", domainId) WHERE 1;`)

		doOneQuery(db, `UPDATE pageInfos SET pageId = CONCAT("zzz", pageId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET lockedBy = CONCAT("zzz", lockedBy) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET seeGroupId = CONCAT("zzz", seeGroupId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET editGroupId = CONCAT("zzz", editGroupId) WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET createdBy = CONCAT("zzz", createdBy) WHERE 1;`)

		doOneQuery(db, `UPDATE pagePairs SET parentId = CONCAT("zzz", parentId) WHERE 1;`)
		doOneQuery(db, `UPDATE pagePairs SET childId = CONCAT("zzz", childId) WHERE 1;`)

		doOneQuery(db, `UPDATE pageSummaries SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE subscriptions SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE subscriptions SET toId = CONCAT("zzz", toId) WHERE 1;`)

		doOneQuery(db, `UPDATE updates SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByPageId = CONCAT("zzz", groupByPageId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByUserId = CONCAT("zzz", groupByUserId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET subscribedToId = CONCAT("zzz", subscribedToId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET goToPageId = CONCAT("zzz", goToPageId) WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET byUserId = CONCAT("zzz", byUserId) WHERE 1;`)

		doOneQuery(db, `UPDATE userMasteryPairs SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE userMasteryPairs SET masteryId = CONCAT("zzz", masteryId) WHERE 1;`)

		doOneQuery(db, `UPDATE users SET id = CONCAT("zzz", id) WHERE 1;`)

		doOneQuery(db, `UPDATE visits SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE visits SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE votes SET userId = CONCAT("zzz", userId) WHERE 1;`)
		doOneQuery(db, `UPDATE votes SET pageId = CONCAT("zzz", pageId) WHERE 1;`)

		doOneQuery(db, `UPDATE pages SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pages SET creatorId = creatorIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE changeLogs SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE changeLogs SET auxPageId = auxPageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE groupMembers SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE groupMembers SET groupId = groupIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE likes SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE likes SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE links SET parentId = parentIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageDomainPairs SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageDomainPairs SET domainId = domainIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageInfos SET pageId = pageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET lockedBy = lockedByBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET seeGroupId = seeGroupIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET editGroupId = editGroupIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pageInfos SET createdBy = createdByBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pagePairs SET parentId = parentIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE pagePairs SET childId = childIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE pageSummaries SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE subscriptions SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE subscriptions SET toId = toIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE updates SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByPageId = groupByPageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET groupByUserId = groupByUserIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET subscribedToId = subscribedToIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET goToPageId = goToPageIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE updates SET byUserId = byUserIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE userMasteryPairs SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE userMasteryPairs SET masteryId = masteryIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE users SET id = idBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE visits SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE visits SET pageId = pageIdBase36 WHERE 1;`)

		doOneQuery(db, `UPDATE votes SET userId = userIdBase36 WHERE 1;`)
		doOneQuery(db, `UPDATE votes SET pageId = pageIdBase36 WHERE 1;`)
		/*
			doOneQuery(db, `ALTER TABLE pages DROP pageIdBase36 ;`)
			doOneQuery(db, `ALTER TABLE pages DROP creatorIdBase36 ;`)

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
		*/
		return "", nil
	})
	if errMessage != "" {
		return 0, err
	}

	return 0, err
}
