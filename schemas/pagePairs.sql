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
		Easy way to memorize: {parentId} is {childId}'s {type}
		Other way to memorize: for each of the relationships you can add
		on the relationship tab of the edit page, the page you're editing
		is the child.*/
	type VARCHAR(32) NOT NULL,

	/* A pair is considered published once its parent and child both have
		pageInfos.currentEdit>0 and not pageInfos.isDeleted. */
	everPublished BOOLEAN NOT NULL,

	UNIQUE(parentId, childId, type),

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
