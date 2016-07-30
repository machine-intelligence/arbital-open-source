/* An entry for every mastery a user knows. */
CREATE TABLE userMasteryPairs (

	/* Id of the user. FK into users. */
	userId VARCHAR(32) NOT NULL,

	/* Id of the mastery. FK into pages. */
	masteryId VARCHAR(32) NOT NULL,

	/* Date this entry was created. */
	createdAt DATETIME NOT NULL,

	/* Date this entry was updated. */
	updatedAt DATETIME NOT NULL,

	/* Set if the user has this mastery. */
	has BOOLEAN NOT NULL,

	/* Set if the user wants to read this. */
	wants BOOLEAN NOT NULL,

	/* Level of understanding. */
	level INT NOT NULL,

	/* Id of the page where the user marked the mastery learned */
	taughtBy VARCHAR(32) NOT NULL,

	/* User's trust when they learned the mastery. FK into userTrustSnapshots */
	trustSnapshotId BIGINT NOT NULL,

	PRIMARY KEY(userId,masteryId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
