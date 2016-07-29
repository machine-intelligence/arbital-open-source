/* This file contains the recent changes to schemas, sorted from oldest to newest. */
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

alter table userMasteryPairs add column level int not null;
