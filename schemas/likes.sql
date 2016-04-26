/* An entry for every like a user cast for a likeable object, such as a page
   or changelog. */
CREATE TABLE likes (
	/* Id of the user who liked. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the likeable this like is for. */
	likeableId BIGINT NOT NULL,

	/* User's trust when they made the like. FK into userTrustSnapshots */
	trustSnapshotId BIGINT NOT NULL,

	/* Like value [-1,1]. */
	value TINYINT NOT NULL,

	/* Date this like was created. */
	createdAt DATETIME NOT NULL,

	/* Date this like was updated. */
	updatedAt DATETIME NOT NULL,

	PRIMARY KEY(userId,likeableId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
