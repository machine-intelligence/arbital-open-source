DROP TABLE IF EXISTS questionTags;

/* This table contains all the question-tag pairs. */
CREATE TABLE questionTags (
	/* Tag id. FK into tags. */
  tagId BIGINT NOT NULL,
	/* Question id. FK into questions. */
  questionId BIGINT NOT NULL,
	/* User id of the person who created this pair. FK into users. */
  createdBy BIGINT NOT NULL,
	/* When this pair was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(tagId, questionId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
