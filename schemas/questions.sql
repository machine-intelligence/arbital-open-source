DROP TABLE IF EXISTS questions;

/* This table contains all the questions we have in our system. */
CREATE TABLE questions (
	/* Unique question id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* User id of the creator of this question. */
  creatorId BIGINT NOT NULL,
	/* Full name of the user who asked. */
	creatorName VARCHAR(64) NOT NULL,
	/* When this question was created. */
  createdAt DATETIME NOT NULL,
	/* Text of the question. */
  text VARCHAR(512) NOT NULL,
	/* Ids for the prior inputs. */
	inputId1 BIGINT NOT NULL,
	inputId2 BIGINT NOT NULL,
	/* Text for the answers. */
	answer1 VARCHAR(32) NOT NULL,
	answer2 VARCHAR(32) NOT NULL,
	/* Privacy key. If not NULL, the question is accessible only with the right link. */
	privacyKey BIGINT,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
