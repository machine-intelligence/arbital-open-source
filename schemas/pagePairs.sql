/* Each row describes a parent-child page relationship. */
CREATE TABLE pagePairs (
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Parent page id. FK into pages. */
  parentId VARCHAR(32) NOT NULL,
	/* Child page id. Part of the FK into pages. */
  childId VARCHAR(32) NOT NULL,
	/* Type of the relationship. 
		parent: parentId is a parent of childId
		tag: parentId is a tag of childId
		requirement: parentId is a requirement of childId
		subject: parentId is a subject that childId teaches
		Easy way to memorize: {parentId} is {childId}'s {type} */
	type VARCHAR(32) NOT NULL,
  UNIQUE(parentId, childId, type),
  PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
