/* This table contains various information about the pages. This info is not
 dependent on any specific edit number. */
CREATE TABLE pageInfos (
	/* Id of the page the info is for. */
	pageId BIGINT NOT NULL,
	/* Edit number currently used to display the page. -1 if this page hasn't
		been published. */
	currentEdit INT NOT NULL,
	/* Maximum edit number used by this page. */
	maxEdit INT NOT NULL,
	/* When this page was originally created. */
	createdAt DATETIME NOT NULL,

	/* If set, the page is locked by this user. FK into users. */
	lockedBy BIGINT NOT NULL,
	/* Time until the user has this lock. */
	lockedUntil DATETIME NOT NULL,
	PRIMARY KEY(pageId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
