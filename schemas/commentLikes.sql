/* An entry for every comment a user liked. */
CREATE TABLE commentLikes (
	/* Id of the user who liked. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the comment this like is for. FK into comments. */
	commentId BIGINT NOT NULL,
	/* Like value [0,1]. */
	value TINYINT NOT NULL,
  /* Date this like was created. */
  createdAt DATETIME NOT NULL,
  /* Date this like was updated. */
  updatedAt DATETIME NOT NULL,
  PRIMARY KEY(userId,commentId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
