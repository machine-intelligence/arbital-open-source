/* A group is a collection of users. */
CREATE TABLE groups (
	/* PK. User's unique id. */
	id BIGINT NOT NULL,
	/* Name of the group. */
	name VARCHAR(64) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	/* Page id for the "welcome" page. */
	rootPageId BIGINT NOT NULL,

	/* === Following are set for domains only. === */
	/* Set to true if this group is a domain. */
	isDomain BOOLEAN NOT NULL,
	/* This is used in the url and as an alias prefix. */
	alias VARCHAR(32) NOT NULL,
	/* Is visible to outsiders. */
	isVisible BOOLEAN NOT NULL,
	UNIQUE(alias),
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
