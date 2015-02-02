DROP TABLE IF EXISTS claimTagPairs;

/* Each row describes a tag that belong to a claim. */
CREATE TABLE claimTagPairs (
	/* Tag id. FK into tags. */
  tagId BIGINT NOT NULL,
	/* Claim id. FK into claims. */
  claimId BIGINT NOT NULL,
	/* User id of the person who created this pair. FK into users. */
  createdBy BIGINT NOT NULL,
	/* When this pair was created. */
  createdAt DATETIME NOT NULL,
  PRIMARY KEY(tagId, claimId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
