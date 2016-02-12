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
