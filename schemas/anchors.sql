/* TODO: we are not actually using this table ATM, but probably should. */
/* This table contains all the anchors. An anchor determines a specific place
	inside a page, including the paragraph and the specific text within it. */
CREATE TABLE anchors (
	/* Id of this anchor. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Text of the paragraph the anchor is in. */
	paragraph MEDIUMTEXT NOT NULL,
	/* Text within the paragraph. If empty, assume it's attached to the entire
		pararaph. */
	text MEDIUMTEXT NOT NULL,
	/* Offset of the text inside the paragraph. */
	offset INT NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
