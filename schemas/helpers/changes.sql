/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table marks add resolvedPageId VARCHAR(32) NOT NULL;
alter table marks add column resolvedBy varchar(32) not null;
alter table marks drop column questionId;
alter table updates add column markId BIGINT NOT NULL;

alter table changeLogs modify column oldSettingsValue varchar(1024) not null;
alter table changeLogs modify column newSettingsValue varchar(1024) not null;

alter table marks add column answered BOOLEAN NOT NULL;
alter table pagePairs add column everPublished boolean not null;
update pagePairs set everPublished = 1
where
	parentId not in (select pageId from pageInfos where currentEdit <= 0) and
	childId not in (select pageId from pageInfos where currentEdit <= 0);

alter table pageInfos add column mergedInto varchar(32) not null;
alter table marks add column type varchar(32) not null;
update marks set type="query";
alter table marks add column resolvedAt datetime not null;
alter table marks add column answeredAt datetime not null;

alter table pageInfos add column likeableId bigint not null;

alter table changeLogs add column likeableId bigint not null;

/* A table for keeping track of likeableIds */
CREATE TABLE likeableIds (
	/* Id of the likeable. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table likes add column likeableId bigint not null;
START TRANSACTION;
SET @likeableId=0;
update pageInfos set likeableId=@likeableId:=@likeableId+1;
insert into likeableIds (id) select likeableId from pageInfos;
update likes join pageInfos on pageInfos.pageId=likes.pageId set likes.likeableId=pageInfos.likeableId;
COMMIT;
alter table likes drop primary key, add primary key (userId, likeableId);
alter table likes drop column pageId;

alter table subscriptions drop column userTrustSnapshotId;

drop table invites;
CREATE TABLE invites (
	/* Id of user sending invite. FK into users. */
	fromUserId VARCHAR(32) NOT NULL,
	/* Id of domain that this invite is for. FK into pageInfos. */
	domainId VARCHAR(32) NOT NULL,
	/* Email address to send invite to. */
	toEmail VARCHAR(100) NOT NULL,
	/* Date this invite was created. */
	createdAt DATETIME NOT NULL,
	/* If a user claimed this invite, this is their id. FK into users. */
	toUserId VARCHAR(32) NOT NULL,
	/* Date this invite was claimed */
	claimedAt DATETIME NOT NULL,
	/* When the invite email was sent. */
	emailSentAt DATETIME NOT NULL,

	PRIMARY KEY(fromUserId,domainId,toEmail)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
DROP TABLE inviteEmailPairs;
drop table userDomainBonusTrust;

ALTER TABLE users ADD COLUMN isTrusted BOOLEAN NOT NULL;

DELETE FROM updates USING updates, pageInfos AS pi WHERE pi.pageId = updates.goToPageId
	AND pi.seeGroupId != '' AND pi.seeGroupId NOT IN (SELECT groupId FROM groupMembers AS gm WHERE gm.userId = updates.userId);

alter table users drop column inviteCode;
alter table users drop column karma;
alter table pageInfos drop column editKarmaLock;

alter table visits add column sessionId VARCHAR(32) NOT NULL after userId;
alter table visits add column ipAddress VARCHAR(64) NOT NULL after sessionId;
delete from groupMembers where userId=groupId;
