/* This table contains all the marks. A mark is an annotation that's attached
	to a page at a specific location. */
CREATE TABLE marks (
	/* Id of this mark. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Type of this mark. */
	type VARCHAR(32) NOT NULL,
	/* Id of the page this mark is on. FK into pageInfos. */
	pageId VARCHAR(32) NOT NULL,
	/* Which edit was live when this mark was created. */
	edit INT NOT NULL,
	/* Id of the user who created this mark. FK into users. */
	creatorId VARCHAR(32) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	/* User's snapshotted requisites. FK into userRequisitePairSnapshots. */
	requisiteSnapshotId BIGINT NOT NULL,
	/* Optional text associated with this mark. */
	text MEDIUMTEXT NOT NULL,

	/* Text of the paragraph the anchor is in. */
	anchorContext MEDIUMTEXT NOT NULL,
	/* Text the comment is attached to. */
	anchorText MEDIUMTEXT NOT NULL,
	/* Offset of the text into the context. */
	anchorOffset INT NOT NULL,

	/* If an author resolves this mark, this variable is set. It might be set to
		the page the mark is on (e.g. for "typo" marks) or to the question page id
		to which the mark was linked (e.g. for "query" marks).
		FK into pageInfos. */
	resolvedPageId VARCHAR(32) NOT NULL,
	/* Id of the user who resolved / dismissed the mark. FK into users. */
	resolvedBy VARCHAR(32) NOT NULL,
	/* When this mark was resolved. */
	resolvedAt DATETIME NOT NULL,

	/* ============= Variables specifically for query-typed marks ============= */
	/* Set to true once there is an answer that works for the given user. */
	answered BOOLEAN NOT NULL,
	/* When this mark was answered. */
	answeredAt DATETIME NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
