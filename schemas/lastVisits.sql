/* Each row is a page-user pair with the date and time when the user has last seen the page. */
CREATE TABLE lastVisits (

	/* FK into users. */
	userId VARCHAR(64) NOT NULL,

	/* Page id. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Date of the first visit. */
	createdAt DATETIME NOT NULL,

	/* Date of the last visit. */
	updatedAt DATETIME NOT NULL,

	UNIQUE(userId,pageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
