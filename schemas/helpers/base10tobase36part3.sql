/* This file contains the second batch of queries used for converting the ids from base10 to base36. */

ALTER TABLE `pages` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pages` ADD `creatorIdBase10` MEDIUMTEXT NOT NULL AFTER `creatorId`;
ALTER TABLE `pages` ADD `privacyKeyBase10` MEDIUMTEXT NOT NULL AFTER `privacyKey`;

ALTER TABLE `changeLogs` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `changeLogs` ADD `auxPageIdBase10` MEDIUMTEXT NOT NULL AFTER `auxPageId`;

ALTER TABLE `fixedIds` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `groupMembers` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `groupMembers` ADD `groupIdBase10` MEDIUMTEXT NOT NULL AFTER `groupId`;

ALTER TABLE `likes` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `likes` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `links` ADD `parentIdBase10` MEDIUMTEXT NOT NULL AFTER `parentId`;

ALTER TABLE `pageDomainPairs` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pageDomainPairs` ADD `domainIdBase10` MEDIUMTEXT NOT NULL AFTER `domainId`;

ALTER TABLE `pageInfos` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `pageInfos` ADD `lockedByBase10` MEDIUMTEXT NOT NULL AFTER `lockedBy`;
ALTER TABLE `pageInfos` ADD `seeGroupIdBase10` MEDIUMTEXT NOT NULL AFTER `seeGroupId`;
ALTER TABLE `pageInfos` ADD `editGroupIdBase10` MEDIUMTEXT NOT NULL AFTER `editGroupId`;
ALTER TABLE `pageInfos` ADD `createdByBase10` MEDIUMTEXT NOT NULL AFTER `createdBy`;

ALTER TABLE `pagePairs` ADD `parentIdBase10` MEDIUMTEXT NOT NULL AFTER `parentId`;
ALTER TABLE `pagePairs` ADD `childIdBase10` MEDIUMTEXT NOT NULL AFTER `childId`;

ALTER TABLE `pageSummaries` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `subscriptions` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `subscriptions` ADD `toIdBase10` MEDIUMTEXT NOT NULL AFTER `toId`;

ALTER TABLE `updates` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `updates` ADD `groupByPageIdBase10` MEDIUMTEXT NOT NULL AFTER `groupByPageId`;
ALTER TABLE `updates` ADD `groupByUserIdBase10` MEDIUMTEXT NOT NULL AFTER `groupByUserId`;
ALTER TABLE `updates` ADD `subscribedToIdBase10` MEDIUMTEXT NOT NULL AFTER `subscribedToId`;
ALTER TABLE `updates` ADD `goToPageIdBase10` MEDIUMTEXT NOT NULL AFTER `goToPageId`;
ALTER TABLE `updates` ADD `byUserIdBase10` MEDIUMTEXT NOT NULL AFTER `byUserId`;

ALTER TABLE `userMasteryPairs` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `userMasteryPairs` ADD `masteryIdBase10` MEDIUMTEXT NOT NULL AFTER `masteryId`;

ALTER TABLE `users` ADD `idBase10` MEDIUMTEXT NOT NULL AFTER `id`;

ALTER TABLE `visits` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `visits` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;

ALTER TABLE `votes` ADD `userIdBase10` MEDIUMTEXT NOT NULL AFTER `userId`;
ALTER TABLE `votes` ADD `pageIdBase10` MEDIUMTEXT NOT NULL AFTER `pageId`;






UPDATE pages SET pageIdBase10 = pageId WHERE 1;
UPDATE pages SET creatorIdBase10 = creatorId WHERE 1;
UPDATE pages SET privacyKeyBase10 = privacyKey WHERE 1;

UPDATE changeLogs SET pageIdBase10 = pageId WHERE 1;
UPDATE changeLogs SET auxPageIdBase10 = auxPageId WHERE 1;

UPDATE fixedIds SET pageIdBase10 = pageId WHERE 1;

UPDATE groupMembers SET userIdBase10 = userId WHERE 1;
UPDATE groupMembers SET groupIdBase10 = groupId WHERE 1;

UPDATE likes SET userIdBase10 = userId WHERE 1;
UPDATE likes SET pageIdBase10 = pageId WHERE 1;

UPDATE links SET parentIdBase10 = parentId WHERE 1;

UPDATE pageDomainPairs SET pageIdBase10 = pageId WHERE 1;
UPDATE pageDomainPairs SET domainIdBase10 = domainId WHERE 1;

UPDATE pageInfos SET pageIdBase10 = pageId WHERE 1;
UPDATE pageInfos SET lockedByBase10 = lockedBy WHERE 1;
UPDATE pageInfos SET seeGroupIdBase10 = seeGroupId WHERE 1;
UPDATE pageInfos SET editGroupIdBase10 = editGroupId WHERE 1;
UPDATE pageInfos SET createdByBase10 = createdBy WHERE 1;

UPDATE pagePairs SET parentIdBase10 = parentId WHERE 1;
UPDATE pagePairs SET childIdBase10 = childId WHERE 1;

UPDATE pageSummaries SET pageIdBase10 = pageId WHERE 1;

UPDATE subscriptions SET userIdBase10 = userId WHERE 1;
UPDATE subscriptions SET toIdBase10 = toId WHERE 1;

UPDATE updates SET userIdBase10 = userId WHERE 1;
UPDATE updates SET groupByPageIdBase10 = groupByPageId WHERE 1;
UPDATE updates SET groupByUserIdBase10 = groupByUserId WHERE 1;
UPDATE updates SET subscribedToIdBase10 = subscribedToId WHERE 1;
UPDATE updates SET goToPageIdBase10 = goToPageId WHERE 1;
UPDATE updates SET byUserIdBase10 = byUserId WHERE 1;

UPDATE userMasteryPairs SET userIdBase10 = userId WHERE 1;
UPDATE userMasteryPairs SET masteryIdBase10 = masteryId WHERE 1;

UPDATE users SET idBase10 = id WHERE 1;

UPDATE visits SET userIdBase10 = userId WHERE 1;
UPDATE visits SET pageIdBase10 = pageId WHERE 1;

UPDATE votes SET userIdBase10 = userId WHERE 1;
UPDATE votes SET pageIdBase10 = pageId WHERE 1;





ALTER TABLE  `pages` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `pages` CHANGE  `creatorId`  `creatorId` VARCHAR( 32 ) NOT NULL ;

ALTER TABLE  `changeLogs` CHANGE  `pageId`  `pageId` VARCHAR( 32 ) NOT NULL ;
ALTER TABLE  `changeLogs` CHANGE  `auxPageId`  `auxPageId` VARCHAR( 32 ) NOT NULL ;

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

UPDATE `changeLogs` SET `pageId` = "" WHERE `pageId` = "0";
UPDATE `changeLogs` SET `auxPageId` = "" WHERE `auxPageId` = "0";

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

ALTER TABLE `changeLogs` ADD `pageIdProcessed` BOOLEAN NOT NULL AFTER `pageId`;
ALTER TABLE `changeLogs` ADD `auxPageIdProcessed` BOOLEAN NOT NULL AFTER `auxPageId`;

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

ALTER TABLE `changeLogs` ADD `pageIdBase36` MEDIUMTEXT NOT NULL AFTER `pageId`;
ALTER TABLE `changeLogs` ADD `auxPageIdBase36` MEDIUMTEXT NOT NULL AFTER `auxPageId`;

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






INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM pages WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT creatorId, createdAt FROM pages WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM changeLogs WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT auxPageId, createdAt FROM changeLogs WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM groupMembers WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT groupId, createdAt FROM groupMembers WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM likes WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM likes WHERE 1;

INSERT INTO pagesandusers (base10id) SELECT parentId FROM links WHERE 1;

INSERT INTO pagesandusers (base10id) SELECT pageId FROM pageDomainPairs WHERE 1;
INSERT INTO pagesandusers (base10id) SELECT domainId FROM pageDomainPairs WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM pageInfos WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT lockedBy, createdAt FROM pageInfos WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT seeGroupId, createdAt FROM pageInfos WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT editGroupId, createdAt FROM pageInfos WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT createdBy, createdAt FROM pageInfos WHERE 1;

INSERT INTO pagesandusers (base10id) SELECT parentId FROM pagePairs WHERE 1;
INSERT INTO pagesandusers (base10id) SELECT childId FROM pagePairs WHERE 1;

INSERT INTO pagesandusers (base10id) SELECT pageId FROM pageSummaries WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM subscriptions WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT toId, createdAt FROM subscriptions WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM updates WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT groupByPageId, createdAt FROM updates WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT groupByUserId, createdAt FROM updates WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT subscribedToId, createdAt FROM updates WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT goToPageId, createdAt FROM updates WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT byUserId, createdAt FROM updates WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM userMasteryPairs WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT masteryId, createdAt FROM userMasteryPairs WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT id, createdAt FROM users WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM visits WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM visits WHERE 1;

INSERT INTO pagesandusers (base10id, createdAt) SELECT userId, createdAt FROM votes WHERE 1;
INSERT INTO pagesandusers (base10id, createdAt) SELECT pageId, createdAt FROM votes WHERE 1;





SELECT pageId FROM pages WHERE pageIdProcessed = 0 UNION
SELECT creatorId FROM pages WHERE creatorIdProcessed = 0 UNION

SELECT pageId FROM changeLogs WHERE pageIdProcessed = 0 UNION
SELECT auxPageId FROM changeLogs WHERE auxPageIdProcessed = 0 UNION

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









SELECT pageId FROM pages WHERE pageIdBase36 = "" UNION
SELECT creatorId FROM pages WHERE creatorIdBase36 = "" UNION

SELECT pageId FROM changeLogs WHERE pageIdBase36 = "" UNION
SELECT auxPageId FROM changeLogs WHERE auxPageIdBase36 = "" UNION

SELECT userId FROM groupMembers WHERE userIdBase36 = "" UNION
SELECT groupId FROM groupMembers WHERE groupIdBase36 = "" UNION

SELECT userId FROM likes WHERE userIdBase36 = "" UNION
SELECT pageId FROM likes WHERE pageIdBase36 = "" UNION

SELECT parentId FROM links WHERE parentIdBase36 = "" UNION

SELECT pageId FROM pageDomainPairs WHERE pageIdBase36 = "" UNION
SELECT domainId FROM pageDomainPairs WHERE domainIdBase36 = "" UNION

SELECT pageId FROM pageInfos WHERE pageIdBase36 = "" UNION
SELECT lockedBy FROM pageInfos WHERE lockedByBase36 = "" UNION
SELECT seeGroupId FROM pageInfos WHERE seeGroupIdBase36 = "" UNION
SELECT editGroupId FROM pageInfos WHERE editGroupIdBase36 = "" UNION
SELECT createdBy FROM pageInfos WHERE createdByBase36 = "" UNION

SELECT parentId FROM pagePairs WHERE parentIdBase36 = "" UNION
SELECT childId FROM pagePairs WHERE childIdBase36 = "" UNION

SELECT pageId FROM pageSummaries WHERE pageIdBase36 = "" UNION

SELECT userId FROM subscriptions WHERE userIdBase36 = "" UNION
SELECT toId FROM subscriptions WHERE toIdBase36 = "" UNION

SELECT userId FROM updates WHERE userIdBase36 = "" UNION
SELECT groupByPageId FROM updates WHERE groupByPageIdBase36 = "" UNION
SELECT groupByUserId FROM updates WHERE groupByUserIdBase36 = "" UNION
SELECT subscribedToId FROM updates WHERE subscribedToIdBase36 = "" UNION
SELECT goToPageId FROM updates WHERE goToPageIdBase36 = "" UNION
SELECT byUserId FROM updates WHERE byUserIdBase36 = "" UNION

SELECT userId FROM userMasteryPairs WHERE userIdBase36 = "" UNION
SELECT masteryId FROM userMasteryPairs WHERE masteryIdBase36 = "" UNION

SELECT id FROM users WHERE idBase36 = "" UNION

SELECT userId FROM visits WHERE userIdBase36 = "" UNION
SELECT pageId FROM visits WHERE pageIdBase36 = "" UNION

SELECT userId FROM votes WHERE userIdBase36 = "" UNION
SELECT pageId FROM votes WHERE pageIdBase36 = "" ;






SELECT pageId FROM pages WHERE pageIdBase36 != pageId UNION
SELECT creatorId FROM pages WHERE creatorIdBase36 != creatorId UNION

SELECT pageId FROM changeLogs WHERE pageIdBase36 != pageId UNION
SELECT auxPageId FROM changeLogs WHERE auxPageIdBase36 != auxPageId UNION

SELECT userId FROM groupMembers WHERE userIdBase36 != userId UNION
SELECT groupId FROM groupMembers WHERE groupIdBase36 != groupId UNION

SELECT userId FROM likes WHERE userIdBase36 != userId UNION
SELECT pageId FROM likes WHERE pageIdBase36 != pageId UNION

SELECT parentId FROM links WHERE parentIdBase36 != parentId UNION

SELECT pageId FROM pageDomainPairs WHERE pageIdBase36 != pageId UNION
SELECT domainId FROM pageDomainPairs WHERE domainIdBase36 != domainId UNION

SELECT pageId FROM pageInfos WHERE pageIdBase36 != pageId UNION
SELECT lockedBy FROM pageInfos WHERE lockedByBase36 != lockedBy UNION
SELECT seeGroupId FROM pageInfos WHERE seeGroupIdBase36 != seeGroupId UNION
SELECT editGroupId FROM pageInfos WHERE editGroupIdBase36 != editGroupId UNION
SELECT createdBy FROM pageInfos WHERE createdByBase36 != createdBy UNION

SELECT parentId FROM pagePairs WHERE parentIdBase36 != parentId UNION
SELECT childId FROM pagePairs WHERE childIdBase36 != childId UNION

SELECT pageId FROM pageSummaries WHERE pageIdBase36 != pageId UNION

SELECT userId FROM subscriptions WHERE userIdBase36 != userId UNION
SELECT toId FROM subscriptions WHERE toIdBase36 != toId UNION

SELECT userId FROM updates WHERE userIdBase36 != userId UNION
SELECT groupByPageId FROM updates WHERE groupByPageIdBase36 != groupByPageId UNION
SELECT groupByUserId FROM updates WHERE groupByUserIdBase36 != groupByUserId UNION
SELECT subscribedToId FROM updates WHERE subscribedToIdBase36 != subscribedToId UNION
SELECT goToPageId FROM updates WHERE goToPageIdBase36 != goToPageId UNION
SELECT byUserId FROM updates WHERE byUserIdBase36 != byUserId UNION

SELECT userId FROM userMasteryPairs WHERE userIdBase36 != userId UNION
SELECT masteryId FROM userMasteryPairs WHERE masteryIdBase36 != masteryId UNION

SELECT id FROM users WHERE idBase36 != id UNION

SELECT userId FROM visits WHERE userIdBase36 != userId UNION
SELECT pageId FROM visits WHERE pageIdBase36 != pageId UNION

SELECT userId FROM votes WHERE userIdBase36 != userId UNION
SELECT pageId FROM votes WHERE pageIdBase36 != pageId ;






SELECT * FROM pages WHERE pageIdBase36 = ""
SELECT * FROM pages WHERE creatorIdBase36 = ""

SELECT * FROM changeLogs WHERE pageIdBase36 = ""
SELECT * FROM changeLogs WHERE auxPageIdBase36 = ""

SELECT * FROM groupMembers WHERE userIdBase36 = ""
SELECT * FROM groupMembers WHERE groupIdBase36 = ""

SELECT * FROM likes WHERE userIdBase36 = ""
SELECT * FROM likes WHERE pageIdBase36 = ""

SELECT * FROM links WHERE parentIdBase36 = ""

SELECT * FROM pageDomainPairs WHERE pageIdBase36 = ""
SELECT * FROM pageDomainPairs WHERE domainIdBase36 = ""

SELECT * FROM pageInfos WHERE pageIdBase36 = ""
SELECT * FROM pageInfos WHERE lockedByBase36 = ""
SELECT * FROM pageInfos WHERE seeGroupIdBase36 = ""
SELECT * FROM pageInfos WHERE editGroupIdBase36 = ""
SELECT * FROM pageInfos WHERE createdByBase36 = ""

SELECT * FROM pagePairs WHERE parentIdBase36 = ""
SELECT * FROM pagePairs WHERE childIdBase36 = ""

SELECT * FROM pageSummaries WHERE pageIdBase36 = ""

SELECT * FROM subscriptions WHERE userIdBase36 = ""
SELECT * FROM subscriptions WHERE toIdBase36 = ""

SELECT * FROM updates WHERE userIdBase36 = ""
SELECT * FROM updates WHERE groupByPageIdBase36 = ""
SELECT * FROM updates WHERE groupByUserIdBase36 = ""
SELECT * FROM updates WHERE subscribedToIdBase36 = ""
SELECT * FROM updates WHERE goToPageIdBase36 = ""
SELECT * FROM updates WHERE byUserIdBase36 = ""

SELECT * FROM userMasteryPairs WHERE userIdBase36 = ""
SELECT * FROM userMasteryPairs WHERE masteryIdBase36 = ""

SELECT * FROM users WHERE idBase36 = ""

SELECT * FROM visits WHERE userIdBase36 = ""
SELECT * FROM visits WHERE pageIdBase36 = ""

SELECT * FROM votes WHERE userIdBase36 = ""
SELECT * FROM votes WHERE pageIdBase36 = ""




SELECT base10id, base36id FROM base10tobase36 WHERE base10id IN (

8639103000879599414,
8639103000879599414,
4213693741839491939,
7722661858289734773,
3158562585659930031,
6820582940749120623,
5534008569097047764,
6053065048861201341,
3560540392275264633,
3560540392275264633,
8138584842800103864,
5092144177314150382,
8992241719442104138,
5933317145970853046,
4675907493088898985,
8676677094741262267,

3440973961008233681,

8992241719442104138,
7648631253816709800,
6686682198220623534,
1407630090992422901
)



