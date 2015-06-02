/* This table contains various information about the pages. This info is not
 dependent on any specific edit number. */
CREATE TABLE pageInfos (
	/* Id of the page the info is for. */
	pageId BIGINT NOT NULL,
	/* Edit number currently used to display the page. */
	currentEdit INT NOT NULL,
	/* Maximum edit number used by this page. */
	maxEdit INT NOT NULL,
	/* When this page was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(pageId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
