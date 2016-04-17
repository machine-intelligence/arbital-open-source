/* This table contains all the marks. A mark is an annotation that's attached
	to a page at a specific location. */
CREATE TABLE marks (
	/* Id of this mark. */
	id BIGINT NOT NULL AUTO_INCREMENT,
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

	/* If the mark is associated to some page, this is the id of that page. This
		can happen if the user picked a page or an author resolved the mark.
		FK into pageInfos. */
	resolvedPageId VARCHAR(32) NOT NULL,
	/* Id of the user who resolved the mark. FK into users. */
	resolvedBy VARCHAR(32) NOT NULL,
	/* Set to true once there is an answer that works for the given user. */
	answered BOOLEAN NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
