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
