/* This table contains all domain pairs that are friendly with each other. */
CREATE TABLE domainFriends (
	/* Domain id. FK into domains. */
	domainId BIGINT NOT NULL,
	/* Id of another domain this domain is friends with. FK into domains. */
	friendId BIGINT NOT NULL,
	/* When this friendship was originally created. */
	createdAt DATETIME NOT NULL,
	/* Id of the user who created the friendship. FK into users. */
	createdBy VARCHAR(32) NOT NULL,

	UNIQUE(domainId,friendId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
