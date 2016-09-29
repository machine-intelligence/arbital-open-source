/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table userMasteryPairs add column level int not null;
update userMasteryPairs set level=1 where has;

update pagePairs set level=4 where level=3;
update pagePairs set level=3 where level=2;
update pagePairs set level=2 where level=1;
update pagePairs set level=1 where level=0;

alter table visits add column analyticsId VARCHAR(64) NOT NULL after sessionId;

CREATE TABLE projects (
	/* Project id. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* The page which describes this project. FK into pages. */
	projectPageId VARCHAR(32) NOT NULL,

	/* The first page the reader should go to. FK into pages. */
	startPageId VARCHAR(32) NOT NULL,

	/* State of the project. "finished", "inProgress", or "requested" */
	state VARCHAR(32) NOT NULL,

	/* Id by which we track who wants to read this. FK into likes. */
	readLikeableId BIGINT NOT NULL,

	/* Id by which we track who wants to write this. FK into likes. */
	writeLikeableId BIGINT NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
insert into projects (projectPageId,startPageId,state,createdAt) values ("5wy","4f4","inProgress",now());
alter table visits modify userId VARCHAR(64) NOT NULL;
alter table lenses add column lensSubtitle VARCHAR(256) NOT NULL;
