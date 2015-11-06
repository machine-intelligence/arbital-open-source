/* This file contains the recent changes to schemas, sorted from oldest to newest. */
alter table pages drop column parents;
drop table groups;
alter table users add column emailFrequency int not null;
alter table users add column emailThreshold int not null;
alter table pages add column editGroupId BIGINT not null;
update pages set sortChildrenBy="recentFirst" where sortChildrenBy="chronological";
alter table pages drop column prevEdit;
CREATE TABLE changeLogs (
	/* Unique update id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* The user who caused this event. FK into users. */
	userId BIGINT NOT NULL,
	/* The affected page. FK into pages. */
	pageId BIGINT NOT NULL,
	/* Edit number of the affected page. Partial FK into pages. */
	edit INT NOT NULL,
	/* Type of update */
	type VARCHAR(32) NOT NULL,
	/* When this update was created. */
	createdAt DATETIME NOT NULL,

	/* This is set for various events. E.g. if a new parent is added, this will
	be set to the parent id. */
	auxPageId BIGINT NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
update pages set editGroupId=creatorId where type="comment";
ALTER TABLE `users` CHANGE `emailFrequency` `emailFrequency` VARCHAR( 16 ) NOT NULL ;
UPDATE `users` SET `emailFrequency`="Daily",`emailThreshold`=3 WHERE `emailThreshold`=0;
UPDATE `users` SET `emailFrequency`="Never" WHERE `emailFrequency`="0";
UPDATE `users` SET `emailFrequency`="Weekly" WHERE `emailFrequency`="1";
UPDATE `users` SET `emailFrequency`="Daily" WHERE `emailFrequency`="2";
UPDATE `users` SET `emailFrequency`="Immediately" WHERE `emailFrequency`="3";
