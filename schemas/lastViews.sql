/* A table for keeping track of the last time the user saw various things */
CREATE TABLE lastViews (
	/* Id of the likeable. */
	userId varchar(32) NOT NULL,
	lastAchievementsView datetime NOT NULL,
	lastReadModeView datetime NOT NULL,
	PRIMARY KEY(userId)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;