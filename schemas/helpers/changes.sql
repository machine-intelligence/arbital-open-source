/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table updates drop column groupByPageId;
alter table updates drop column groupByUserId;

/* This table contains what pages belong to which paths. */
CREATE TABLE pathPages (
	/* Id of this entry. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page guide that starts this path. FK into pageInfos. */
	guideId VARCHAR(32) NOT NULL,
	/* Id of one of the pages on the path. FK into pageInfos. */
	pathPageId VARCHAR(32) NOT NULL,
	/* Ordering index when ordering the pages in a path. */
	pathIndex INT NOT NULL,
	/* Id of the user who created the relationship. FK into users. */
	createdBy VARCHAR(32) NOT NULL,
	/* When this lens relationship was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the last user who updated the relationship. FK into users. */
	updatedBy VARCHAR(32) NOT NULL,
	/* When this relationship was updated last. */
	updatedAt DATETIME NOT NULL,

	/* This constraint should apply, but makes it very difficult to update the index for multiple rows */
	/* UNIQUE(guideId,index),*/
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE pathInstances (
	/* Id of this entry. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* User who started this path. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the page guide that started this path. FK into pageInfos. */
	guideId VARCHAR(32) NOT NULL,
	/* Comma separated list of page ids which this path has. FK into pageInfos. */
	pageIds TEXT NOT NULL,
	/* Index of the page the user is on. */
	progress INT NOT NULL,
	/* When this instance was created. */
	createdAt DATETIME NOT NULL,
	/* When this instance was updated last. */
	updatedAt DATETIME NOT NULL,
	/* Optional. If set, the user copied the path from this instance. */
	originalInstanceId BIGINT NOT NULL,
	/* Set to true when the user finished the path. */
	isFinished BOOLEAN NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table users add column isSlackMember boolean not null;
alter table pathInstances add column sourcePageIds text not null;
alter table invites add column bonusEditTrust int not null;

alter table pagePairs add column level int not null;
alter table pagePairs add column isStrong boolean not null;

CREATE TABLE userTrust (
	/* The user this trust corresponds to. FK into users. */
	userId varchar(32) NOT NULL,
	/* The domain that these trust scores belong to. */
	domainId varchar(32) NOT NULL,
	/* The user's trust for general actions */
	generalTrust INT NOT NULL,
	/* The user's trust for editing actions */
	editTrust INT NOT NULL,

	UNIQUE(userId,domainId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
insert into userTrust (userId,domainId,editTrust) select toUserId,domainId,200+bonusEditTrust from invites;
alter table invites drop column bonusEditTrust;

alter table pageInfos add column viewCount BIGINT NOT NULL;
update pageInfos as pi set pi.viewCount=(select count(distinct userId) from visits as v where v.pageId=pi.pageId);

/* An entry for every content request pair (page and type) */
CREATE TABLE contentRequests (

	/* Id of the request. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* The page the request was made for. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Type of request. Either slowDown or speedUp */
	type VARCHAR(32) NOT NULL,

	/* Id by which we track likes. FK into likes. */
	likeableId BIGINT NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	/* There can only be one row per (page, type) pair */
	UNIQUE KEY(pageId,type),

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table visits change column sessionId sessionId VARCHAR(64);

update pagePairs set level=1 where level=2;
update pagePairs set level=2 where level=3 OR level=4;
update pagePairs set level=3 where level=5;
