DROP TABLE IF EXISTS updates;

/* This table contains all the updates we have in our system. */
CREATE TABLE updates (
	/* Unique update id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* User id of the owner of this update. FK into users. */
  userId BIGINT NOT NULL,
	/* Id of the question the update is for. FK into questions. */
  questionId BIGINT NOT NULL,
	/* Id of the comment the update is for. FK into comments. */
  commentId BIGINT NOT NULL,
	/* Type of update */
	type VARCHAR(32) NOT NULL,
	/* When this update was created. */
  createdAt DATETIME NOT NULL,
	/* When this was last updated. */
	updatedAt DATETIME NOT NULL,
	/* Number of such updates. */
	count INT NOT NULL,
	/* True iff the user has seen these updates. While false, we can continue to stack similar updates together. */
	seen BOOLEAN NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
