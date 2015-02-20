DROP TABLE IF EXISTS answers;

/* Each row is an answer to a page that's of type "question". */
CREATE TABLE answers (
	/* Page id. FK into pages. */
  pageId BIGINT NOT NULL,
	/* Index id of this answer within the context of the quesiton. */
	indexId SMALLINT NOT NULL,
	/* Text for this answer. */
	text VARCHAR(64) NOT NULL,
	/* When this answer was created. */
  createdAt DATETIME NOT NULL,
	/* When this answer was updated. */
	updatedAt DATETIME NOT NULL,
  PRIMARY KEY(pageId, indexId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
