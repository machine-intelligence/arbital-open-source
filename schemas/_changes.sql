/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table updates add column byUserId bigint not null;
alter table pages add column metaText mediumtext not null;
alter table pages drop column privacyKey;
alter table pages drop column deletedBy;
alter table updates add column emailed boolean not null;
ALTER TABLE pages CHANGE COLUMN `karmaLock` `editKarmaLock` INT NOT NULL;
ALTER TABLE pages CHANGE COLUMN `groupId` `seeGroupId` BIGINT NOT NULL;
ALTER TABLE pagePairs ADD COLUMN type VARCHAR(32) NOT NULL;
update pagePairs set type="parent";
drop index parentId on pagePairs;
ALTER TABLE pagePairs ADD UNIQUE (parentId,childId,type);
CREATE TABLE userMasteryPairs (
	/* Id of the user. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the mastery. FK into pages. */
	masteryId BIGINT NOT NULL,
  /* Date this entry was created. */
  createdAt DATETIME NOT NULL,
  /* Date this entry was updated. */
  updatedAt DATETIME NOT NULL,
	/* Set if the user has this mastery. */
	has BOOLEAN NOT NULL,
	/* Set if the user manually set the 'has' value. */
	isManuallySet BOOLEAN NOT NULL,
  PRIMARY KEY(userId,masteryId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
alter table pages drop column parents;
drop table groups;
alter table users add column emailFrequency int not null;
alter table users add column emailThreshold int not null;
UPDATE `users` SET `emailFrequency`=2,`emailThreshold`=3 WHERE `emailThreshold`=0;
alter table pages add column editGroupId BIGINT not null;