/* A group is a collection of users. */
CREATE TABLE groups (
	/* PK. User's unique id. */
	id BIGINT NOT NULL,
	/* Name of the group. */
	name VARCHAR(64) NOT NULL,
	/* When this was created. */
	createdAt DATETIME NOT NULL,
	UNIQUE(name),
	PRIMARY KEY(id)
) CHARACTER SET utf8 COLLATE utf8_general_ci;
