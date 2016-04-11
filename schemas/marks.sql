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
	/* If this mark is associated with a question, this is the id. FK into pageInfos. */
	questionId VARCHAR(32) NOT NULL,
	/* User's snapshotted requisites. FK into userRequisitePairSnapshots. */
	requisiteSnapshotId BIGINT NOT NULL;

	/* Text of the paragraph the anchor is in. */
	anchorContext MEDIUMTEXT NOT NULL,
	/* Text the comment is attached to. */
	anchorText MEDIUMTEXT NOT NULL,
	/* Offset of the text into the context. */
	anchorOffset INT NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
