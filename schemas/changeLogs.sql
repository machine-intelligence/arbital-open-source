/* This table contains an entry for every change that a page undergoes. */
CREATE TABLE changeLogs (

	/* Unique update id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* Likeable id for this changelog. Partial FK into likes.
	   Note that this is not set until the first time this changelog is liked. */
	likeableId BIGINT NOT NULL,

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

	/* So that we can show what changed in the change log. */
	oldSettingsValue VARCHAR(1024) NOT NULL,
	newSettingsValue VARCHAR(1024) NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
