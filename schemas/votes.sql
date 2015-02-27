DROP TABLE IF EXISTS votes;

/* An entry for every probability vote a user casts for a question. There could be
 multiple votes from one user for the same page. */
CREATE TABLE votes (
  /* PK. Vote's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the user who voted. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the page this vote is for. FK into pages. */
	pageId BIGINT NOT NULL,
	/* Vote value. */
	value DOUBLE NOT NULL,
  /* Date this like was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
