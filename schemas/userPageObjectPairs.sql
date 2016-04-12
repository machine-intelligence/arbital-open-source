/* This table contains an entry for each (user, page object) pair, where
	the object can store some user specific data. For example, multiple choice
	questions can store the user's answer. */
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
