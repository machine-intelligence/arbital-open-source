// base10ToBase36Part1Task.go does part 1 of converting all the ids from base 10 to base 36
package tasks

import (
	"fmt"
	//"regexp"
	//"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/user"
)

// Base10ToBase36Part1Task is the object that's put into the daemon queue.
type Base10ToBase36Part1Task struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *Base10ToBase36Part1Task) IsValid() error {
	return nil
}

var lastBase36Id string

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *Base10ToBase36Part1Task) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Debugf("==== PART 1 START ====")
	defer c.Debugf("==== PART 1 COMPLETED ====")

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		doOneQuery(db, `ALTER TABLE  pages CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pages CHANGE  creatorId  creatorId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pages CHANGE  privacyKey  privacyKey VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  changeLogs CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  changeLogs CHANGE  auxPageId  auxPageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  fixedIds CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  groupMembers CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  groupMembers CHANGE  groupId  groupId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  likes CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  likes CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  links CHANGE  parentId  parentId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  pageDomainPairs CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pageDomainPairs CHANGE  domainId  domainId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  pageInfos CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pageInfos CHANGE  lockedBy  lockedBy VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pageInfos CHANGE  seeGroupId  seeGroupId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pageInfos CHANGE  editGroupId  editGroupId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pageInfos CHANGE  createdBy  createdBy VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  pagePairs CHANGE  parentId  parentId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  pagePairs CHANGE  childId  childId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  pageSummaries CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  subscriptions CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  subscriptions CHANGE  toId  toId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  updates CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  updates CHANGE  groupByPageId  groupByPageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  updates CHANGE  groupByUserId  groupByUserId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  updates CHANGE  subscribedToId  subscribedToId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  updates CHANGE  goToPageId  goToPageId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  updates CHANGE  byUserId  byUserId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  userMasteryPairs CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  userMasteryPairs CHANGE  masteryId  masteryId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  users CHANGE  id  id VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  visits CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  visits CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `ALTER TABLE  votes CHANGE  userId  userId VARCHAR( 32 ) NOT NULL ;`)
		doOneQuery(db, `ALTER TABLE  votes CHANGE  pageId  pageId VARCHAR( 32 ) NOT NULL ;`)

		doOneQuery(db, `UPDATE pages SET pageId = "" WHERE pageId = "0";`)
		doOneQuery(db, `UPDATE pages SET creatorId = "" WHERE creatorId = "0";`)
		doOneQuery(db, `UPDATE pages SET privacyKey = "" WHERE privacyKey = "0";`)

		doOneQuery(db, `UPDATE changeLogs SET pageId = "" WHERE pageId = "0";`)
		doOneQuery(db, `UPDATE changeLogs SET auxPageId = "" WHERE auxPageId = "0";`)

		doOneQuery(db, `UPDATE fixedIds SET pageId = "" WHERE pageId = "0";`)

		doOneQuery(db, `UPDATE groupMembers SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE groupMembers SET groupId = "" WHERE groupId = "0";`)

		doOneQuery(db, `UPDATE likes SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE likes SET pageId = "" WHERE pageId = "0";`)

		doOneQuery(db, `UPDATE links SET parentId = "" WHERE parentId = "0";`)

		doOneQuery(db, `UPDATE pageDomainPairs SET pageId = "" WHERE pageId = "0";`)
		doOneQuery(db, `UPDATE pageDomainPairs SET domainId = "" WHERE domainId = "0";`)

		doOneQuery(db, `UPDATE pageInfos SET pageId = "" WHERE pageId = "0";`)
		doOneQuery(db, `UPDATE pageInfos SET lockedBy = "" WHERE lockedBy = "0";`)
		doOneQuery(db, `UPDATE pageInfos SET seeGroupId = "" WHERE seeGroupId = "0";`)
		doOneQuery(db, `UPDATE pageInfos SET editGroupId = "" WHERE editGroupId = "0";`)
		doOneQuery(db, `UPDATE pageInfos SET createdBy = "" WHERE createdBy = "0";`)

		doOneQuery(db, `UPDATE pagePairs SET parentId = "" WHERE parentId = "0";`)
		doOneQuery(db, `UPDATE pagePairs SET childId = "" WHERE childId = "0";`)

		doOneQuery(db, `UPDATE pageSummaries SET pageId = "" WHERE pageId = "0";`)

		doOneQuery(db, `UPDATE subscriptions SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE subscriptions SET toId = "" WHERE toId = "0";`)

		doOneQuery(db, `UPDATE updates SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE updates SET groupByPageId = "" WHERE groupByPageId = "0";`)
		doOneQuery(db, `UPDATE updates SET groupByUserId = "" WHERE groupByUserId = "0";`)
		doOneQuery(db, `UPDATE updates SET subscribedToId = "" WHERE subscribedToId = "0";`)
		doOneQuery(db, `UPDATE updates SET goToPageId = "" WHERE goToPageId = "0";`)
		doOneQuery(db, `UPDATE updates SET byUserId = "" WHERE byUserId = "0";`)

		doOneQuery(db, `UPDATE userMasteryPairs SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE userMasteryPairs SET masteryId = "" WHERE masteryId = "0";`)

		doOneQuery(db, `UPDATE users SET id = "" WHERE id = "0";`)

		doOneQuery(db, `UPDATE visits SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE visits SET pageId = "" WHERE pageId = "0";`)

		doOneQuery(db, `UPDATE votes SET userId = "" WHERE userId = "0";`)
		doOneQuery(db, `UPDATE votes SET pageId = "" WHERE pageId = "0";`)

		doOneQuery(db, `ALTER TABLE pages ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pages ADD creatorIdProcessed BOOLEAN NOT NULL AFTER creatorId;`)
		doOneQuery(db, `ALTER TABLE pages ADD privacyKeyProcessed BOOLEAN NOT NULL AFTER privacyKey;`)

		doOneQuery(db, `ALTER TABLE changeLogs ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE changeLogs ADD auxPageIdProcessed BOOLEAN NOT NULL AFTER auxPageId;`)

		doOneQuery(db, `ALTER TABLE fixedIds ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE groupMembers ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE groupMembers ADD groupIdProcessed BOOLEAN NOT NULL AFTER groupId;`)

		doOneQuery(db, `ALTER TABLE likes ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE likes ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE links ADD parentIdProcessed BOOLEAN NOT NULL AFTER parentId;`)

		doOneQuery(db, `ALTER TABLE pageDomainPairs ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pageDomainPairs ADD domainIdProcessed BOOLEAN NOT NULL AFTER domainId;`)

		doOneQuery(db, `ALTER TABLE pageInfos ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD lockedByProcessed BOOLEAN NOT NULL AFTER lockedBy;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD seeGroupIdProcessed BOOLEAN NOT NULL AFTER seeGroupId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD editGroupIdProcessed BOOLEAN NOT NULL AFTER editGroupId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD createdByProcessed BOOLEAN NOT NULL AFTER createdBy;`)

		doOneQuery(db, `ALTER TABLE pagePairs ADD parentIdProcessed BOOLEAN NOT NULL AFTER parentId;`)
		doOneQuery(db, `ALTER TABLE pagePairs ADD childIdProcessed BOOLEAN NOT NULL AFTER childId;`)

		doOneQuery(db, `ALTER TABLE pageSummaries ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE subscriptions ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE subscriptions ADD toIdProcessed BOOLEAN NOT NULL AFTER toId;`)

		doOneQuery(db, `ALTER TABLE updates ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE updates ADD groupByPageIdProcessed BOOLEAN NOT NULL AFTER groupByPageId;`)
		doOneQuery(db, `ALTER TABLE updates ADD groupByUserIdProcessed BOOLEAN NOT NULL AFTER groupByUserId;`)
		doOneQuery(db, `ALTER TABLE updates ADD subscribedToIdProcessed BOOLEAN NOT NULL AFTER subscribedToId;`)
		doOneQuery(db, `ALTER TABLE updates ADD goToPageIdProcessed BOOLEAN NOT NULL AFTER goToPageId;`)
		doOneQuery(db, `ALTER TABLE updates ADD byUserIdProcessed BOOLEAN NOT NULL AFTER byUserId;`)

		doOneQuery(db, `ALTER TABLE userMasteryPairs ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE userMasteryPairs ADD masteryIdProcessed BOOLEAN NOT NULL AFTER masteryId;`)

		doOneQuery(db, `ALTER TABLE users ADD idProcessed BOOLEAN NOT NULL AFTER id;`)

		doOneQuery(db, `ALTER TABLE visits ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE visits ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE votes ADD userIdProcessed BOOLEAN NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE votes ADD pageIdProcessed BOOLEAN NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE pages ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pages ADD creatorIdBase36 MEDIUMTEXT NOT NULL AFTER creatorId;`)
		doOneQuery(db, `ALTER TABLE pages ADD privacyKeyBase36 MEDIUMTEXT NOT NULL AFTER privacyKey;`)

		doOneQuery(db, `ALTER TABLE changeLogs ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE changeLogs ADD auxPageIdBase36 MEDIUMTEXT NOT NULL AFTER auxPageId;`)

		doOneQuery(db, `ALTER TABLE fixedIds ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE groupMembers ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE groupMembers ADD groupIdBase36 MEDIUMTEXT NOT NULL AFTER groupId;`)

		doOneQuery(db, `ALTER TABLE likes ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE likes ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE links ADD parentIdBase36 MEDIUMTEXT NOT NULL AFTER parentId;`)

		doOneQuery(db, `ALTER TABLE pageDomainPairs ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pageDomainPairs ADD domainIdBase36 MEDIUMTEXT NOT NULL AFTER domainId;`)

		doOneQuery(db, `ALTER TABLE pageInfos ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD lockedByBase36 MEDIUMTEXT NOT NULL AFTER lockedBy;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD seeGroupIdBase36 MEDIUMTEXT NOT NULL AFTER seeGroupId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD editGroupIdBase36 MEDIUMTEXT NOT NULL AFTER editGroupId;`)
		doOneQuery(db, `ALTER TABLE pageInfos ADD createdByBase36 MEDIUMTEXT NOT NULL AFTER createdBy;`)

		doOneQuery(db, `ALTER TABLE pagePairs ADD parentIdBase36 MEDIUMTEXT NOT NULL AFTER parentId;`)
		doOneQuery(db, `ALTER TABLE pagePairs ADD childIdBase36 MEDIUMTEXT NOT NULL AFTER childId;`)

		doOneQuery(db, `ALTER TABLE pageSummaries ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE subscriptions ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE subscriptions ADD toIdBase36 MEDIUMTEXT NOT NULL AFTER toId;`)

		doOneQuery(db, `ALTER TABLE updates ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE updates ADD groupByPageIdBase36 MEDIUMTEXT NOT NULL AFTER groupByPageId;`)
		doOneQuery(db, `ALTER TABLE updates ADD groupByUserIdBase36 MEDIUMTEXT NOT NULL AFTER groupByUserId;`)
		doOneQuery(db, `ALTER TABLE updates ADD subscribedToIdBase36 MEDIUMTEXT NOT NULL AFTER subscribedToId;`)
		doOneQuery(db, `ALTER TABLE updates ADD goToPageIdBase36 MEDIUMTEXT NOT NULL AFTER goToPageId;`)
		doOneQuery(db, `ALTER TABLE updates ADD byUserIdBase36 MEDIUMTEXT NOT NULL AFTER byUserId;`)

		doOneQuery(db, `ALTER TABLE userMasteryPairs ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE userMasteryPairs ADD masteryIdBase36 MEDIUMTEXT NOT NULL AFTER masteryId;`)

		doOneQuery(db, `ALTER TABLE users ADD idBase36 MEDIUMTEXT NOT NULL AFTER id;`)

		doOneQuery(db, `ALTER TABLE visits ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE visits ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)

		doOneQuery(db, `ALTER TABLE votes ADD userIdBase36 MEDIUMTEXT NOT NULL AFTER userId;`)
		doOneQuery(db, `ALTER TABLE votes ADD pageIdBase36 MEDIUMTEXT NOT NULL AFTER pageId;`)

		doOneQuery(db, `DROP TABLE pagesandusers;`)
		doOneQuery(db, `DROP TABLE base10tobase36;`)

		doOneQuery(db, `
CREATE TABLE pagesandusers (
	base10id VARCHAR(32) NOT NULL,
	createdAt DATETIME NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;
`)

		doOneQuery(db, `
CREATE TABLE base10tobase36 (
	base10id VARCHAR(32) NOT NULL,
	createdAt DATETIME NOT NULL,
	base36id VARCHAR(32) NOT NULL
) CHARACTER SET utf8 COLLATE utf8_general_ci;
`)

		doOneQuery(db, `
INSERT INTO pagesandusers (base10id, createdAt)
SELECT pageId, createdAt
FROM pages
WHERE 1;
`)

		doOneQuery(db, `
INSERT INTO pagesandusers (base10id, createdAt)
SELECT id, createdAt
FROM users
WHERE 1;
`)

		doOneQuery(db, `
INSERT INTO base10tobase36 (base10id, createdAt)
SELECT DISTINCT pagesandusers.base10id, pagesandusers.createdAt
FROM pagesandusers
INNER JOIN
    (SELECT pagesandusers.base10id, MIN(pagesandusers.createdAt) AS minCreatedAt
    FROM pagesandusers
    GROUP BY pagesandusers.base10id) groupedpagesandusers
ON pagesandusers.base10id = groupedpagesandusers.base10id
AND pagesandusers.createdAt = groupedpagesandusers.minCreatedAt;
`)

		lastBase36Id = "0"
		rows := db.NewStatement(`
				SELECT base10id,createdAt
				FROM base10tobase36
				WHERE 1
				ORDER BY createdAt
				`).Query()
		if err = rows.Process(oneRowUpdateBase10ToBase36); err != nil {
			c.Debugf("ERROR, failed to fix text: %v", err)
			return "", err
		}

		doOneQuery(db, `DELETE FROM base10tobase36 WHERE base36id=""`)

		return "", nil
	})
	if errMessage != "" {
		return 0, err
	}

	return 0, err
}

func doOneQuery(db *database.DB, queryString string) error {
	//db.C.Debugf("queryString: %v", queryString)
	statement := db.NewStatement(queryString)
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't execute query "+queryString+": %v", err)
	}
	statement.Close()

	return nil
}

func oneRowUpdateBase10ToBase36(db *database.DB, rows *database.Rows) error {
	var base10Id, createdAt string
	if err := rows.Scan(&base10Id, &createdAt); err != nil {
		return fmt.Errorf("failed to scan a page: %v", err)
	}

	//db.C.Debugf("lastBase36Id: %v", lastBase36Id)
	//db.C.Debugf("base10Id: %v", base10Id)
	//db.C.Debugf("createdAt: %v", createdAt)
	base36Id, err := user.IncrementBase31Id(db, lastBase36Id)
	if err != nil {
		return fmt.Errorf("Error incrementing id: %v", err)
	}
	db.C.Debugf("base36Id: %v", base36Id)
	lastBase36Id = base36Id

	hashmap := make(map[string]interface{})
	hashmap["base10id"] = base10Id
	hashmap["createdAt"] = createdAt
	hashmap["base36id"] = base36Id
	statement := db.NewInsertStatement("base10tobase36", hashmap, "base36id")
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update base10tobase36 table: %v", err)
	}

	return nil
}
