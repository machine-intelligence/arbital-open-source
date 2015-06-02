/* A comment is a short piece of text submitted by a user for some page. */
CREATE TABLE comments (
	/* Unique comment id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the page this comment belongs to. FK into pages. */
	pageId BIGINT NOT NULL,
	/* Id of the comment this is a reply to. 0 if it's top level. */
	replyToId BIGINT NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	/* When this was last updated. */
	updatedAt DATETIME NOT NULL,
	/* User id of the creator. */
	creatorId BIGINT NOT NULL,
	/* Text supplied by the user. */
	text VARCHAR(2048) NOT NULL,
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
