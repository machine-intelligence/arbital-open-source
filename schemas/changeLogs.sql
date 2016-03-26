/* This table contains an entry for every change that a page undergoes. */
CREATE TABLE changeLogs (

	/* Unique update id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* The user who caused this event. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* The affected page. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Edit number of the affected page. Partial FK into pages. */
	edit INT NOT NULL,

	/* Type of update */
	type VARCHAR(32) NOT NULL,

	/* When this update was created. */
	createdAt DATETIME NOT NULL,

	/* This is set for various events. E.g. if a new parent is added, this will
	be set to the parent id. */
	auxPageId VARCHAR(32) NOT NULL,

	/* For use in the change log and on the updates page. */
	oldSettingsValue VARCHAR(32) NOT NULL,
	newSettingsValue VARCHAR(32) NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
