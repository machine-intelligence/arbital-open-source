/* This file contains the recent changes to schemas, sorted from oldest to newest. */
ALTER TABLE `users` CHANGE `emailFrequency` `emailFrequency` VARCHAR( 16 ) NOT NULL ;
UPDATE `users` SET `emailFrequency`="daily",`emailThreshold`=3 WHERE `emailThreshold`=0;
UPDATE `users` SET `emailFrequency`="never" WHERE `emailFrequency`="0";
UPDATE `users` SET `emailFrequency`="weekly" WHERE `emailFrequency`="1";
UPDATE `users` SET `emailFrequency`="daily" WHERE `emailFrequency`="2";
UPDATE `users` SET `emailFrequency`="immediately" WHERE `emailFrequency`="3";
/* November 10 */
alter table pageInfos add column alias VARCHAR(64) NOT NULL,
	add column type VARCHAR(32) NOT NULL,
	add column sortChildrenBy VARCHAR(32) NOT NULL,
	add column hasVote BOOLEAN NOT NULL,
	add column voteType VARCHAR(32) NOT NULL,
	add column seeGroupId BIGINT NOT NULL,
	add column editGroupId BIGINT NOT NULL,
	add column editKarmaLock INT NOT NULL;
update pageInfos as pi join pages as p on (pi.pageId=p.pageId and p.isCurrentEdit) set pi.alias=p.alias,pi.type=p.type,pi.sortChildrenBy=p.sortChildrenBy,pi.hasVote=p.hasVote,pi.voteType=p.voteType,pi.seeGroupId=p.seeGroupId,pi.editGroupId=p.editGroupId,pi.editKarmaLock=p.editKarmaLock ;
alter table pages drop column `alias`,drop column `type`,drop column sortChildrenBy,drop column hasVote,drop column voteType,drop column seeGroupId,drop column editGroupId,drop column editKarmaLock;
update subscriptions set toPageId=toUserId where toPageId=0; alter table subscriptions drop column toUserId; alter table subscriptions change column toPageId toId bigint not null;
update updates set subscribedToPageId=subscribedToUserId where subscribedToPageId=0; alter table updates drop column subscribedToUserId; alter table updates change column subscribedToPageId subscribedToId bigint not null;
