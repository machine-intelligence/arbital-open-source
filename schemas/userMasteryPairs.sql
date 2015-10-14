/* An entry for every mastery a user knows. */
CREATE TABLE userMasteryPairs (
	/* Id of the user. FK into users. */
	userId BIGINT NOT NULL,
	/* Id of the mastery. FK into pages. */
	masteryId BIGINT NOT NULL,
  /* Date this entry was created. */
  createdAt DATETIME NOT NULL,
  /* Date this entry was updated. */
  updatedAt DATETIME NOT NULL,
	/* Set if the user has this mastery. */
	has BOOLEAN NOT NULL,
	/* Set if the user manually set the 'has' value. */
	isManuallySet BOOLEAN NOT NULL,
  PRIMARY KEY(userId,masteryId)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
