DROP TABLE IF EXISTS commentVotes;

/* An entry for every vote a user cast for a comment. */
CREATE TABLE commentVotes (
	/* Id of the user who voted. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the comment this vote is for. FK into comments. */
	commentId BIGINT NOT NULL,
	/* Vote value [0,1]. */
	value TINYINT NOT NULL,
  /* Date this vote was created. */
  createdAt DATETIME NOT NULL,
  /* Date this vote was updated. */
  updatedAt DATETIME NOT NULL,
  PRIMARY KEY(userId,commentId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
