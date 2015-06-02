/* This table contains all the aliases for pages. An alias is a short string
 that points to a specific page. It's of the form "name-suffix", where
 'name' is a user specified string, and 'suffix' is a number of make full name
 unique.
 Ideally, we *never* want to delete aliases, since each one could potentially
 still be in use somewhere. */
CREATE TABLE aliases (
	/* Unique alias name. */
	fullName VARCHAR(64) NOT NULL,
	/* Non-unique standardized name. E.g. for "Vitamin_D_Good-2", this will be "vitamindgood". */
	standardizedName VARCHAR(64) NOT NULL,
	/* Number appended to the name to make it unique. */
	suffix SMALLINT NOT NULL,
	/* Id of the page the alias points to. */
	pageId BIGINT NOT NULL,
	/* User id of the creator of this alias. */
	creatorId BIGINT NOT NULL,
	/* When this alias was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(fullName)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
