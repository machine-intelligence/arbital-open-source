/* An entry for every red link. */
CREATE TABLE redLinks (
	/* Alias of the red link. */
	alias VARCHAR(64) NOT NULL,

	/* Id by which we track requests. Partial FK into likes. */
	likeableId BIGINT NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	PRIMARY KEY(alias)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
