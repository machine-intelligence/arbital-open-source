DROP TABLE IF EXISTS votes;

/* An entry for every vote a user cast for a claim There could be
multiple votes from one user for the same claim. */
CREATE TABLE votes (
  /* PK. Vote's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the user who voted. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the claim this vote is for. FK into claims. */
	claimId BIGINT NOT NULL,
	/* Vote value [-1,1]. */
	value TINYINT NOT NULL,
  /* Date this vote was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
