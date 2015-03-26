DROP TABLE IF EXISTS pagePairs;

/* Each row describes a parent-child page relationship. */
CREATE TABLE pagePairs (
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Parent page id. FK into pages. */
  parentId BIGINT NOT NULL,
	/* Child page id. Part of the FK into pages. */
  childId BIGINT NOT NULL,
	/* Child page edit number. Part of the FK into pages. */
	childEdit INT NOT NULL,
  UNIQUE(parentId, childId, childEdit),
  PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
