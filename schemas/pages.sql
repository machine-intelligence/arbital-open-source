DROP TABLE IF EXISTS pages;

/* This table contains all the edits for all the pages, including the original edit.
 Each row is one edit for a given page. */
CREATE TABLE pages (
	/* Id of the page the edit is for. */
	pageId BIGINT NOT NULL,
	/* The edit number. The current version always has edit=0. Older versions have
	 higher edit numbers. E.g. the previous version has edit=1. */
	edit INT NOT NULL,
	/* Page's type. */
	type VARCHAR(32) NOT NULL,
	/* User id of the creator of this edit. */
	creatorId BIGINT NOT NULL,
	/* When this edit was created. */
	createdAt DATETIME NOT NULL,
	/* Title of the page. */
	title VARCHAR(512) NOT NULL,
	/* Text of the page. */
	text MEDIUMTEXT NOT NULL,
	/* True iff the page has a probability vote. */
	hasVote BOOLEAN NOT NULL,
	/* Minimum amount of karma a user needs to edit this page. */
	karmaLock INT NOT NULL,
	/* If > 0, the page is accessible only with the right link. For drafts
	 this is always set. */
	privacyKey BIGINT NOT NULL,
	/* If not 0, this edit has been deleted by this user id. */
	deletedBy BIGINT NOT NULL,
	/* Set to true iff this is a draft. Drafts can only be edited by their
	 creator and they always have the privacy key set. */
	isDraft BOOLEAN NOT NULL,
	PRIMARY KEY(pageId, edit)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
