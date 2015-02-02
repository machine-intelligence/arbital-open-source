DROP TABLE IF EXISTS claims;

/* This table contains all the claims. */
CREATE TABLE claims (
	/* Unique claim id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* User id of the creator of this claim. */
  creatorId BIGINT NOT NULL,
	/* Full name of the user who asked. */
	creatorName VARCHAR(64) NOT NULL,
	/* When this claim was created. */
  createdAt DATETIME NOT NULL,
	/* When this question was last updated. */
  updatedAt DATETIME NOT NULL,
	/* Text of the claim. */
  text VARCHAR(512) NOT NULL,
	/* Link associated with this claim. */
	url VARCHAR(2048) NOT NULL,
	/* Privacy key. If not NULL, the claim is accessible only with the right link. */
	privacyKey BIGINT,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
