/* An entry for every like a user cast for a page. There could be
multiple likes from one user for the same page. */
CREATE TABLE likes (
	/* Id of the user who liked. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the page this like is for. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Like value [-1,1]. */
	value TINYINT NOT NULL,

	/* Date this like was created. */
	createdAt DATETIME NOT NULL,

	/* Date this like was updated. */
	updatedAt DATETIME NOT NULL,

	PRIMARY KEY(userId,pageId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
