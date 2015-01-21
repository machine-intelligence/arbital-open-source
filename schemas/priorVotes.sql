DROP TABLE IF EXISTS priorVotes;

/* An entry for every vote a user cast for a prior. */
CREATE TABLE priorVotes (
  /* PK. Vote's unique id. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the user who voted. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the support this vote is for. FK into support. */
	supportId BIGINT NOT NULL,
	/* Vote value. */
	value FLOAT NOT NULL,
  /* Date this vote was last changed. */
  lastChanged DATETIME NOT NULL,
	CONSTRAINT oneVotePerSupport UNIQUE(userId, supportId),
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
