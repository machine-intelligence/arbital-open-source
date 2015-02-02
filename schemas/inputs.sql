DROP TABLE IF EXISTS inputs;

/* An input is a connection between two claims. Child claim is being used to argue
for or against the parent claim. */
CREATE TABLE inputs (
	/* Unique input id. PK. */
  id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the parent claim. FK into claims. */
  parentId BIGINT NOT NULL,
	/* Id of the child claim. FK into claims. */
  childId BIGINT NOT NULL,
	/* When this was created. */
  createdAt DATETIME NOT NULL,
	/* When this was last updated. */
  updatedAt DATETIME NOT NULL,
	/* User id of the creator. */
  creatorId BIGINT NOT NULL,
	/* Full name of the creator. */
	creatorName VARCHAR(64) NOT NULL,
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
