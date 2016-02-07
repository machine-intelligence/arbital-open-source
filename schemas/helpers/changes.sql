/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table users add column ignoreMathjax bool not null;
alter table pageInfos add column lensIndex int not null;
alter table userMasteryPairs drop column isManuallySet;
alter table userMasteryPairs add column wants boolean not null;
alter table pageInfos add column isRequisite BOOL NOT NULL;
alter table pageInfos add column indirectTeacher BOOL NOT NULL;

ALTER TABLE  `pages` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pages` CHANGE  `creatorId`  `creatorId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pages` CHANGE  `privacyKey`  `privacyKey` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `changeLogs` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `changeLogs` CHANGE  `auxPageId`  `auxPageId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `fixedIds` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `groupMembers` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `groupMembers` CHANGE  `groupId`  `groupId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `likes` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `likes` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `links` CHANGE  `parentId`  `parentId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `pageDomainPairs` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pageDomainPairs` CHANGE  `domainId`  `domainId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `pageInfos` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pageInfos` CHANGE  `lockedBy`  `lockedBy` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pageInfos` CHANGE  `seeGroupId`  `seeGroupId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pageInfos` CHANGE  `editGroupId`  `editGroupId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pageInfos` CHANGE  `createdBy`  `createdBy` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `pagePairs` CHANGE  `parentId`  `parentId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pagePairs` CHANGE  `childId`  `childId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `pageSummaries` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `subscriptions` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `subscriptions` CHANGE  `toId`  `toId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `updates` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `updates` CHANGE  `groupByPageId`  `groupByPageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `updates` CHANGE  `groupByUserId`  `groupByUserId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `updates` CHANGE  `subscribedToId`  `subscribedToId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `updates` CHANGE  `goToPageId`  `goToPageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `updates` CHANGE  `byUserId`  `byUserId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `userMasteryPairs` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `userMasteryPairs` CHANGE  `masteryId`  `masteryId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `users` CHANGE  `id`  `id` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `visits` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `visits` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `votes` CHANGE  `userId`  `userId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `votes` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;




UPDATE `pages` SET `pageId` = "" WHERE `pageId` = "0";
UPDATE `pages` SET `creatorId` = "" WHERE `creatorId` = "0";
UPDATE `pages` SET `privacyKey` = "" WHERE `privacyKey` = "0";

UPDATE `changeLogs` SET `pageId` = "" WHERE `pageId` = "0";
UPDATE `changeLogs` SET `auxPageId` = "" WHERE `auxPageId` = "0";

UPDATE `fixedIds` SET `pageId` = "" WHERE `pageId` = "0";

UPDATE `groupMembers` SET `userId` = "" WHERE `userId` = "0";
UPDATE `groupMembers` SET `groupId` = "" WHERE `groupId` = "0";

UPDATE `likes` SET `userId` = "" WHERE `userId` = "0";
UPDATE `likes` SET `pageId` = "" WHERE `pageId` = "0";

UPDATE `links` SET `parentId` = "" WHERE `parentId` = "0";

UPDATE `pageDomainPairs` SET `pageId` = "" WHERE `pageId` = "0";
UPDATE `pageDomainPairs` SET `domainId` = "" WHERE `domainId` = "0";

UPDATE `pageInfos` SET `pageId` = "" WHERE `pageId` = "0";
UPDATE `pageInfos` SET `lockedBy` = "" WHERE `lockedBy` = "0";
UPDATE `pageInfos` SET `seeGroupId` = "" WHERE `seeGroupId` = "0";
UPDATE `pageInfos` SET `editGroupId` = "" WHERE `editGroupId` = "0";
UPDATE `pageInfos` SET `createdBy` = "" WHERE `createdBy` = "0";

UPDATE `pagePairs` SET `parentId` = "" WHERE `parentId` = "0";
UPDATE `pagePairs` SET `childId` = "" WHERE `childId` = "0";

UPDATE `pageSummaries` SET `pageId` = "" WHERE `pageId` = "0";

UPDATE `subscriptions` SET `userId` = "" WHERE `userId` = "0";
UPDATE `subscriptions` SET `toId` = "" WHERE `toId` = "0";

UPDATE `updates` SET `userId` = "" WHERE `userId` = "0";
UPDATE `updates` SET `groupByPageId` = "" WHERE `groupByPageId` = "0";
UPDATE `updates` SET `groupByUserId` = "" WHERE `groupByUserId` = "0";
UPDATE `updates` SET `subscribedToId` = "" WHERE `subscribedToId` = "0";
UPDATE `updates` SET `goToPageId` = "" WHERE `goToPageId` = "0";
UPDATE `updates` SET `byUserId` = "" WHERE `byUserId` = "0";

UPDATE `userMasteryPairs` SET `userId` = "" WHERE `userId` = "0";
UPDATE `userMasteryPairs` SET `masteryId` = "" WHERE `masteryId` = "0";

UPDATE `users` SET `id` = "" WHERE `id` = "0";

UPDATE `visits` SET `userId` = "" WHERE `userId` = "0";
UPDATE `visits` SET `pageId` = "" WHERE `pageId` = "0";

UPDATE `votes` SET `userId` = "" WHERE `userId` = "0";
UPDATE `votes` SET `pageId` = "" WHERE `pageId` = "0";
