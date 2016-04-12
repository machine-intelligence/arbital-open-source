/* When we snapshot all user's requisites, we store them in this table. Each snapshot
	has the same id, but takes up multiple rows. */
CREATE TABLE userRequisitePairSnapshots (
	/* Id of the snapshot. Note that this is not unique per row. */
	id BIGINT NOT NULL,
	/* Id of the user. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the requisite. FK into pages. */
	requisiteId VARCHAR(32) NOT NULL,
 	/* Date this entry was created. */
 	createdAt DATETIME NOT NULL,
	/* Set if the user has this mastery. */
	has BOOLEAN NOT NULL,
	/* Set if the user wants to read this. */
	wants BOOLEAN NOT NULL
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
