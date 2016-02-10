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


ALTER TABLE `pages` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;
ALTER TABLE `pages` ADD `creatorIdProcessed` BOOLEAN NOT NULL AFTER `creatorId`;
ALTER TABLE `pages` ADD `privacyKeyProcessed` BOOLEAN NOT NULL AFTER `privacyKey`;

ALTER TABLE `changeLogs` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;
ALTER TABLE `changeLogs` ADD `auxPageIdProcessed` BOOLEAN NOT NULL AFTER `auxPageId`;

ALTER TABLE `fixedIds` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;

ALTER TABLE `groupMembers` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `groupMembers` ADD `groupIdProcessed` BOOLEAN NOT NULL AFTER `groupId`;

ALTER TABLE `likes` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `likes` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;

ALTER TABLE `links` ADD `parentIdProcessed` BOOLEAN NOT NULL AFTER `parentId`;

ALTER TABLE `pageDomainPairs` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;
ALTER TABLE `pageDomainPairs` ADD `domainIdProcessed` BOOLEAN NOT NULL AFTER `domainId`;

ALTER TABLE `pageInfos` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;
ALTER TABLE `pageInfos` ADD `lockedByProcessed` BOOLEAN NOT NULL AFTER `lockedBy`;
ALTER TABLE `pageInfos` ADD `seeGroupIdProcessed` BOOLEAN NOT NULL AFTER `seeGroupId`;
ALTER TABLE `pageInfos` ADD `editGroupIdProcessed` BOOLEAN NOT NULL AFTER `editGroupId`;
ALTER TABLE `pageInfos` ADD `createdByProcessed` BOOLEAN NOT NULL AFTER `createdBy`;

ALTER TABLE `pagePairs` ADD `parentIdProcessed` BOOLEAN NOT NULL AFTER `parentId`;
ALTER TABLE `pagePairs` ADD `childIdProcessed` BOOLEAN NOT NULL AFTER `childId`;

ALTER TABLE `pageSummaries` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;

ALTER TABLE `subscriptions` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `subscriptions` ADD `toIdProcessed` BOOLEAN NOT NULL AFTER `toId`;

ALTER TABLE `updates` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `updates` ADD `groupByPageIdProcessed` BOOLEAN NOT NULL AFTER `groupByPageId`;
ALTER TABLE `updates` ADD `groupByUserIdProcessed` BOOLEAN NOT NULL AFTER `groupByUserId`;
ALTER TABLE `updates` ADD `subscribedToIdProcessed` BOOLEAN NOT NULL AFTER `subscribedToId`;
ALTER TABLE `updates` ADD `goToPageIdProcessed` BOOLEAN NOT NULL AFTER `goToPageId`;
ALTER TABLE `updates` ADD `byUserIdProcessed` BOOLEAN NOT NULL AFTER `byUserId`;

ALTER TABLE `userMasteryPairs` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `userMasteryPairs` ADD `masteryIdProcessed` BOOLEAN NOT NULL AFTER `masteryId`;

ALTER TABLE `users` ADD `idProcessed` BOOLEAN NOT NULL AFTER `id`;

ALTER TABLE `visits` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `visits` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;

ALTER TABLE `votes` ADD `userIdProcessed` BOOLEAN NOT NULL AFTER `userId`;
ALTER TABLE `votes` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;






ALTER TABLE `pages` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pages` ADD `creatorIdBase36` MEDIUMTEXT NOT NULL AFTER `creatorId`;
ALTER TABLE `pages` ADD `privacyKeyBase36` MEDIUMTEXT NOT NULL AFTER `privacyKey`;

ALTER TABLE `changeLogs` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `changeLogs` ADD `auxPageIdBase36` MEDIUMTEXT NOT NULL AFTER `auxPageId`;

ALTER TABLE `fixedIds` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `groupMembers` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `groupMembers` ADD `groupIdBase36` MEDIUMTEXT NOT NULL AFTER `groupId`;

ALTER TABLE `likes` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `likes` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `links` ADD `parentIdBase36` MEDIUMTEXT NOT NULL AFTER `parentId`;

ALTER TABLE `pageDomainPairs` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pageDomainPairs` ADD `domainIdBase36` MEDIUMTEXT NOT NULL AFTER `domainId`;

ALTER TABLE `pageInfos` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pageInfos` ADD `lockedByBase36` MEDIUMTEXT NOT NULL AFTER `lockedBy`;
ALTER TABLE `pageInfos` ADD `seeGroupIdBase36` MEDIUMTEXT NOT NULL AFTER `seeGroupId`;
ALTER TABLE `pageInfos` ADD `editGroupIdBase36` MEDIUMTEXT NOT NULL AFTER `editGroupId`;
ALTER TABLE `pageInfos` ADD `createdByBase36` MEDIUMTEXT NOT NULL AFTER `createdBy`;

ALTER TABLE `pagePairs` ADD `parentIdBase36` MEDIUMTEXT NOT NULL AFTER `parentId`;
ALTER TABLE `pagePairs` ADD `childIdBase36` MEDIUMTEXT NOT NULL AFTER `childId`;

ALTER TABLE `pageSummaries` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `subscriptions` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `subscriptions` ADD `toIdBase36` MEDIUMTEXT NOT NULL AFTER `toId`;

ALTER TABLE `updates` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `updates` ADD `groupByPageIdBase36` MEDIUMTEXT NOT NULL AFTER `groupByPageId`;
ALTER TABLE `updates` ADD `groupByUserIdBase36` MEDIUMTEXT NOT NULL AFTER `groupByUserId`;
ALTER TABLE `updates` ADD `subscribedToIdBase36` MEDIUMTEXT NOT NULL AFTER `subscribedToId`;
ALTER TABLE `updates` ADD `goToPageIdBase36` MEDIUMTEXT NOT NULL AFTER `goToPageId`;
ALTER TABLE `updates` ADD `byUserIdBase36` MEDIUMTEXT NOT NULL AFTER `byUserId`;

ALTER TABLE `userMasteryPairs` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `userMasteryPairs` ADD `masteryIdBase36` MEDIUMTEXT NOT NULL AFTER `masteryId`;

ALTER TABLE `users` ADD `idBase36` MEDIUMTEXT NOT NULL AFTER `id`;

ALTER TABLE `visits` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `visits` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `votes` ADD `userIdBase36` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `votes` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;





SELECT pageId FROM pages WHERE pageIdProcessed = 0 UNION
SELECT creatorId FROM pages WHERE creatorIdProcessed = 0 UNION
SELECT privacyKey FROM pages WHERE privacyKeyProcessed = 0 UNION

SELECT pageId FROM changeLogs WHERE pageIdProcessed = 0 UNION
SELECT auxPageId FROM changeLogs WHERE auxPageIdProcessed = 0 UNION

SELECT pageId FROM fixedIds WHERE pageIdProcessed = 0 UNION

SELECT userId FROM groupMembers WHERE userIdProcessed = 0 UNION
SELECT groupId FROM groupMembers WHERE groupIdProcessed = 0 UNION

SELECT userId FROM likes WHERE userIdProcessed = 0 UNION
SELECT pageId FROM likes WHERE pageIdProcessed = 0 UNION

SELECT parentId FROM links WHERE parentIdProcessed = 0 UNION

SELECT pageId FROM pageDomainPairs WHERE pageIdProcessed = 0 UNION
SELECT domainId FROM pageDomainPairs WHERE domainIdProcessed = 0 UNION

SELECT pageId FROM pageInfos WHERE pageIdProcessed = 0 UNION
SELECT lockedBy FROM pageInfos WHERE lockedByProcessed = 0 UNION
SELECT seeGroupId FROM pageInfos WHERE seeGroupIdProcessed = 0 UNION
SELECT editGroupId FROM pageInfos WHERE editGroupIdProcessed = 0 UNION
SELECT createdBy FROM pageInfos WHERE createdByProcessed = 0 UNION

SELECT parentId FROM pagePairs WHERE parentIdProcessed = 0 UNION
SELECT childId FROM pagePairs WHERE childIdProcessed = 0 UNION

SELECT pageId FROM pageSummaries WHERE pageIdProcessed = 0 UNION

SELECT userId FROM subscriptions WHERE userIdProcessed = 0 UNION
SELECT toId FROM subscriptions WHERE toIdProcessed = 0 UNION

SELECT userId FROM updates WHERE userIdProcessed = 0 UNION
SELECT groupByPageId FROM updates WHERE groupByPageIdProcessed = 0 UNION
SELECT groupByUserId FROM updates WHERE groupByUserIdProcessed = 0 UNION
SELECT subscribedToId FROM updates WHERE subscribedToIdProcessed = 0 UNION
SELECT goToPageId FROM updates WHERE goToPageIdProcessed = 0 UNION
SELECT byUserId FROM updates WHERE byUserIdProcessed = 0 UNION

SELECT userId FROM userMasteryPairs WHERE userIdProcessed = 0 UNION
SELECT masteryId FROM userMasteryPairs WHERE masteryIdProcessed = 0 UNION

SELECT id FROM users WHERE idProcessed = 0 UNION

SELECT userId FROM visits WHERE userIdProcessed = 0 UNION
SELECT pageId FROM visits WHERE pageIdProcessed = 0 UNION

SELECT userId FROM votes WHERE userIdProcessed = 0 UNION
SELECT pageId FROM votes WHERE pageIdProcessed = 0 ;

