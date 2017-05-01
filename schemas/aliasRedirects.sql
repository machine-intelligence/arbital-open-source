/* When a page's alias is changed, we add a row in this table. */
CREATE TABLE aliasRedirects (

	/* The old alias. */
	oldAlias VARCHAR(64) NOT NULL,

	/* The new alias. */
	newAlias VARCHAR(64) NOT NULL,

	UNIQUE(oldAlias)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
