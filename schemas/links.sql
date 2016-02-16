/* When a parent page has a link to a child page, we add a row in this table. */
CREATE TABLE links (
	/* Id of the parent page. FK into pages. */
	parentId VARCHAR(32) NOT NULL,
	/* Alias or id of the child claim. */
	childAlias VARCHAR(64) NOT NULL,
	UNIQUE(parentId, childAlias)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
