DROP TABLE IF EXISTS visits;

/* This table contains all the question-user pairs with the date and time of
when the user has last seen the question. */
CREATE TABLE visits (
	/* User id. FK into users. */
  userId BIGINT NOT NULL,
	/* Question id. FK into questions. */
  questionId BIGINT NOT NULL,
	/* When this visit was created. */
  createdAt DATETIME NOT NULL,
	/* When this visit was last updated. */
  updatedAt DATETIME NOT NULL,
  PRIMARY KEY(userId, questionId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
