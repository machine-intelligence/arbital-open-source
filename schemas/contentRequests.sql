/* An entry for every content request pair (page and type) */
CREATE TABLE contentRequests (

	/* Id of the request. */
	id BIGINT NOT NULL AUTO_INCREMENT,

	/* The page the request was made for. FK into pages. */
	pageId VARCHAR(32) NOT NULL,

	/* Type of request. E.g. slowDown, speedUp, etc. */
	type VARCHAR(32) NOT NULL,

	/* Id by which we track likes. FK into likes. */
	likeableId BIGINT NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	/* There can only be one row per (page, type) pair */
	UNIQUE KEY(pageId,type),

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
