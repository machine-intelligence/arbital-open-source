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
