/* An entry for every search string that's attached to a page. */
CREATE TABLE searchStrings (
	/* Id of this search string. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page this string is for. FK into pages. */
	pageId VARCHAR(32) NOT NULL,
	/* Id of the user who added this string. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* String's text. */
	text VARCHAR(1024) NOT NULL,
	/* Date this string was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
