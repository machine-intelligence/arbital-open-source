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

CREATE TABLE invites (
	/* PK: Invite's unique code. */
	code VARCHAR(32) NOT NULL,
	/* Type of invite: personal or group */
	type VARCHAR(32) NOT NULL,
	/* Id of user sending invite. FK into users.*/
	senderId VARCHAR(32) NOT NULL,
	/* Id of domain that this invite is for. FK into pages. */
	domainId VARCHAR(32) NOT NULL,
	/* Date this invite was added to the table. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(code)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE TABLE inviteEmailPairs (
	/* Invite's unique code. FK into invites */
	code VARCHAR(32) NOT NULL,
	/* Email address to send invite to */
	email VARCHAR(100) NOT NULL,
	/* Id of user claiming invite. FK into users */
	claimingUserId VARCHAR(32) NOT NULL,
	/* Date this invite was claimed */
	claimedAt DATETIME NOT NULL,
	PRIMARY KEY(code, email)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE TABLE userDomainBonusTrust (
	/* Id of User. FK into users.*/
	userId VARCHAR(32) NOT NULL,
	/* Id of the domain the page belongs to. FK into groups. */
	domainId VARCHAR(32) NOT NULL,
	/* BonusTrust score a user has to edit pages in this domain */
	bonusEditTrust BIGINT NOT NULL,
	PRIMARY KEY(userId, domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

ALTER TABLE users ADD COLUMN isTrusted BOOLEAN NOT NULL;

DELETE FROM updates USING updates, pageInfos AS pi WHERE pi.pageId = updates.goToPageId
	AND pi.seeGroupId != '' AND pi.seeGroupId NOT IN (SELECT groupId FROM groupMembers AS gm WHERE gm.userId = updates.userId)
