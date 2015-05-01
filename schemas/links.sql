DROP TABLE IF EXISTS links;

/* When a parent page has a link to a child page, we add a row in this table. */
CREATE TABLE links (
	/* Unique link id. PK. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* Id of the parent page. FK into pages. */
	parentId BIGINT NOT NULL,
	/* Id of the child claim. FK into pages. */
	childId BIGINT NOT NULL,
	/* Alias/id of an unlinked child page. Set iff childId is not. */
	unlinkedChildAlias VARCHAR(64) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	UNIQUE(parentId, childId, unlinkedChildAlias),
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
