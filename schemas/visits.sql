DROP TABLE IF EXISTS visits;

/* Each row is a claim-user pair with the date and time of
when the user has last seen the claim. */
CREATE TABLE visits (
	/* User id. FK into users. */
  userId BIGINT NOT NULL,
	/* Claim id. FK into claims. */
  claimId BIGINT NOT NULL,
	/* When this visit was created. */
  createdAt DATETIME NOT NULL,
	/* When this visit was last updated. */
  updatedAt DATETIME NOT NULL,
  PRIMARY KEY(userId, claimId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
