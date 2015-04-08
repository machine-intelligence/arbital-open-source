DROP TABLE IF EXISTS groups;

/* A group is a collection of users. */
CREATE TABLE groups (
	/* Name of the group. */
	name VARCHAR(64) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	PRIMARY KEY(name)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
