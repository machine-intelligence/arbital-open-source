DROP TABLE IF EXISTS support;

/* This table contains all the supporting statements we have in our system. */
CREATE TABLE support (
	/* Unique support id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the question this fact belong to. FK into questions. */
  questionId BIGINT NOT NULL,
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Full name of the creator. */
	creatorName VARCHAR(128) NOT NULL,
	/* Text supplied by the user. */
  text VARCHAR(2048) NOT NULL,
	/* Index of the answer this statement supports. Start at 1. */
  answerIndex TINYINT UNSIGNED NOT NULL,
	/* Prior percent for this answer. */
	prior FLOAT,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
