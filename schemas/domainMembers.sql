/* An entry for every member in a domain. */
CREATE TABLE domainMembers (

	/* Id of the domain. FK into domains. */
	domainId BIGINT NOT NULL,

	/* Id of the user member. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Date this user was added. */
	createdAt DATETIME NOT NULL,

	/* User's role in this domain. */
	role VARCHAR(32) NOT NULL,

	PRIMARY KEY(domainId,userId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
