DROP TABLE IF EXISTS statements;

/* This table contains all the statements we have in our system. */
CREATE TABLE statements (
	/* Unique statement id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* When this statement was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator of this statement. */
  creatorId BIGINT NOT NULL,
	/* Text of the statement. */
  text VARCHAR(512) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
