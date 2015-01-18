DROP TABLE IF EXISTS questions;

/* This table contains all the questions we have in our system. */
CREATE TABLE questions (
	/* Unique question id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* When this question was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator of this question. */
  creatorId BIGINT NOT NULL,
	/* Full name of the user who asked. */
	creatorName VARCHAR(128) NOT NULL,
	/* Text of the question. */
  text VARCHAR(512) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
