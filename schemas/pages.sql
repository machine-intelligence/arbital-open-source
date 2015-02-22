DROP TABLE IF EXISTS pages;

/* This table contains all the edits for all the pages, including the original edit.
 Each row is one edit for a given page. */
CREATE TABLE pages (
	/* Unique id of the edit. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Non-unique. Id of the page the edit is for. */
	pageId BIGINT NOT NULL,
	/* The edit number of this page. Edits start at 0 and go up. */
	edit INT NOT NULL,
	/* Page's type. */
	type VARCHAR(32) NOT NULL,
	/* User id of the creator of this edit. */
	creatorId BIGINT NOT NULL,
	/* Full name of the user who created this edit. */
	creatorName VARCHAR(64) NOT NULL,
	/* When this edit was created. */
	createdAt DATETIME NOT NULL,
	/* Title of the page. */
	title VARCHAR(512) NOT NULL,
	/* Text of the page. */
	text MEDIUMTEXT NOT NULL,
	/* Minimum amount of karma a user needs to edit this page. */
	karmaLock INT NOT NULL,
	/* Privacy key. If not NULL, the page is accessible only with the right link. */
	privacyKey BIGINT,
	/* If not 0, this pages has been deleted by this user id. */
	deletedBy BIGINT NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE INDEX pageIdIndex ON pages (pageId);
