DROP TABLE IF EXISTS priorVotes;

/* An entry for every prior vote a user cast for a question. There could be
multiple votes from one user for the same question. */
CREATE TABLE priorVotes (
  /* PK. Vote's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the user who voted. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the question this vote is for. FK into questions. */
	questionId BIGINT NOT NULL,
	/* Vote value [0,100]. For questions with only two answers, this is the
	 probability assigned to the first answer.*/
	value FLOAT NOT NULL,
  /* Date this vote was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
