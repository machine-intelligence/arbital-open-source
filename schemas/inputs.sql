DROP TABLE IF EXISTS inputs;

/* This table contains all the input statements we have in our system. */
CREATE TABLE inputs (
	/* Unique input id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the question this input belongs to. FK into questions. */
  questionId BIGINT NOT NULL,
	/* What kind of input this is. */
	/*inputType VARCHAR(32) NOT NULL,*/
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* When this was last updated. */
  updatedAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Full name of the creator. */
	creatorName VARCHAR(64) NOT NULL,
	/* Description supplied by the user. */
  text VARCHAR(2048) NOT NULL,
	/* Link to a study / paper. */
	url VARCHAR(2048) NOT NULL,
	/* The kind of study the link points to. */
	/* studyType VARCHAR(64),*/
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
