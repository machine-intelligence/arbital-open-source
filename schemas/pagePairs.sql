DROP TABLE IF EXISTS pagePairs;

/* Each row describes a parent-child page relationship. */
CREATE TABLE pagePairs (
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Parent page id. FK into pages. */
  parentId BIGINT NOT NULL,
	/* Child page id. Part of the FK into pages. */
  childId BIGINT NOT NULL,
	/* Child page edit number. Part of the FK into pages. This is not set if
	 userId is set. */
	childEdit INT NOT NULL,
	/* Optional. Id of the user who owns this personal pairing. */
	userId BIGINT NOT NULL,
  UNIQUE(parentId, childId, childEdit, userId),
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
