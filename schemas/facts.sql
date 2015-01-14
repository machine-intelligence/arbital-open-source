DROP TABLE IF EXISTS facts;

/* This table contains all the facts we have in our system. */
CREATE TABLE facts (
	/* Unique fact id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the statement this fact belong to. FK into statements. */
  statementId BIGINT NOT NULL,
	/* When this fact was created. */
  createdAt DATETIME NOT NULL,
	/* User id of the creator of this fact. */
  creatorId BIGINT NOT NULL,
	/* Text of the fact. */
  text VARCHAR(2048) NOT NULL,
	/* True iff this is a fact of support. False iff for opposition. */
	isSupport BOOLEAN NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
