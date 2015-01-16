DROP TABLE IF EXISTS support;

/* This table contains all the supporting statements we have in our system. */
CREATE TABLE support (
	/* Unique support id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the question this fact belong to. FK into questions. */
  questionId BIGINT NOT NULL,
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Text supplied by the user. */
  text VARCHAR(2048) NOT NULL,
	/* What answer this statement supports. */
  answer VARCHAR(512) NOT NULL,
	/* Prior percent for this answer. */
	prior SMALLINT,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
