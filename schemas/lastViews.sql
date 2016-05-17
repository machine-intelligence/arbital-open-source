/* A table for keeping track of the last time the user saw various things */
CREATE TABLE lastViews (
	/* Id of the likeable. */
	userId varchar(32) NOT NULL,
	/* The thing the user saw. */
	viewName varchar(64) NOT NULL,
	/* The last time the user viewed the thing. */
	viewedAt DATETIME NOT NULL,
	PRIMARY KEY(userId,viewName)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
