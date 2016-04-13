/* This file contains the recent changes to schemas, sorted from oldest to newest. */
CREATE TABLE userPageObjectPairs (
	/* Id of the user the user is for. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the page the info is for. */
	pageId VARCHAR(32) NOT NULL,
	/* Page's published edit at the time this value was set. */
	edit INT NOT NULL,
	/* Alias name of the object. */
	object VARCHAR(64) NOT NULL,
	/* When this value was originally created at. */
	createdAt DATETIME NOT NULL,
	/* When this value was updated. */
	updatedAt DATETIME NOT NULL,

	/* Whatever value the object decides to set here. */
	value VARCHAR(512) NOT NULL,

	PRIMARY KEY(userId,pageId,object)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
alter table pageInfos add column isEditorComment bool not null;
alter table pages add column prevEdit int not null;
create table copyPages like pages;
insert copyPages select * from pages;
update pages as p set prevEdit=(select max(cp.edit) from copyPages as cp where p.pageId=cp.pageId and NOT cp.isAutosave and NOT cp.isSnapshot and cp.createdAt<p.createdAt);
drop table copyPages;
alter table pages add column snapshotText mediumtext not null;

alter table changeLogs add column oldSettingsValue varchar(32) not null;
alter table changeLogs add column newSettingsValue varchar(32) not null;

delete from changeLogs where type = '';
delete changeLogs from changeLogs join pageInfos on changeLogs.auxPageId=pageInfos.pageId where pageInfos.currentEdit <= 0;
update changeLogs set type = 'newTeacher' where type = 'newTeaches';
update updates set type = 'newTeacher' where type = 'newTeaches';

alter table updates add column changeLogId varchar(32) not null;
alter table updates add column unseen bool not null;
update updates set unseen = TRUE where newCount > 0;
alter table updates drop column newCount;

update changeLogs set type = 'newRequiredBy' where type = 'newRequiredFor';
update changeLogs set type = 'deleteRequiredBy' where type = 'deleteRequiredFor';
update changeLogs set type = 'newUsedAsTag' where type = 'newTagTarget';
update changeLogs set type = 'deleteUsedAsTag' where type = 'deleteTagTarget';
update updates set type = 'newUsedAsTag' where type = 'newTaggedBy';
update updates set type = 'deleteUsedAsTag' where type = 'deleteTaggedBy';

alter table userMasteryPairs add column taughtBy varchar(32) not null;

alter table pages change isCurrentEdit `isLiveEdit` BOOLEAN NOT NULL;
alter table pageInfos add column isDeleted BOOLEAN NOT NULL;
update pageInfos,
(select
	pageId,
	count(*) > 0 as ever_published,
	max(edit) as last_published_edit,
	sum(isLiveEdit) > 0 as is_live
	from pages
	where not isSnapshot and not isAutosave
	group by pageId
	having ever_published and not is_live)
as deleted_pages
set
	pageInfos.currentEdit = deleted_pages.last_published_edit,
	pageInfos.isDeleted = true
where pageInfos.pageId = deleted_pages.pageId;

alter table changeLogs modify oldSettingsValue varchar(64) not null;
alter table changeLogs modify newSettingsValue varchar(64) not null;

CREATE TABLE marks (
	/* Id of this mark. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page this mark is on. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Which edit was live when this mark was created. */
	edit INT NOT NULL,
	/* Id of the user who created this mark. FK into users. */
	creatorId VARCHAR(32) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	/* If this mark is associated with a question, this is the id. FK into pageInfos. */
	questionId VARCHAR(32) NOT NULL,

	/* Text of the paragraph the anchor is in. */
	anchorContext MEDIUMTEXT NOT NULL,
	/* Text the comment is attached to. */
	anchorText MEDIUMTEXT NOT NULL,
	/* Offset of the text into the context. */
	anchorOffset INT NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

delete from likes where id not in (select max(id) from (select * from likes) as newestLikes group by userId,pageId);
alter table likes modify id bigint, drop primary key, add primary key(userId, pageId);
alter table likes drop column id;
alter table likes add column updatedAt datetime not null;
update likes set updatedAt=createdAt;

ALTER TABLE marks ADD COLUMN requisiteSnapshotId BIGINT NOT NULL;

CREATE TABLE userRequisitePairSnapshots (
	/* Id of the snapshot. Note that this is not unique per row. */
	id BIGINT NOT NULL,
	/* Id of the user. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the requisite. FK into pages. */
	requisiteId VARCHAR(32) NOT NULL,
 	/* Date this entry was created. */
 	createdAt DATETIME NOT NULL,
	/* Set if the user has this mastery. */
	has BOOLEAN NOT NULL,
	/* Set if the user wants to read this. */
	wants BOOLEAN NOT NULL
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table marks add column text mediumtext not null;

CREATE TABLE searchStrings (
	/* Id of this search string. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page this string is for. FK into pages. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the user who added this string. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* String's text. */
	text VARCHAR(1024) NOT NULL,
	/* Date this string was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE answers (
	/* Id of this answer. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the question this answer is for. FK into pages. */
	questionId VARCHAR(32) NOT NULL,
	/* Id of the user who added this string. FK into users. */
	answerPageId VARCHAR(32) NOT NULL,
	/* Id of the user who added this answer. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Date this answer was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

alter table marks add resolvedPageId VARCHAR(32) NOT NULL;
