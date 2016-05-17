/* A table for keeping track of the last time the user saw various things */
CREATE TABLE lastViews (
	/* Id of the user who viewed the thing. */
	userId varchar(32) NOT NULL,
	/* The thing the user viewed. */
	viewName varchar(64) NOT NULL,
	/* The last time the user viewed the thing. */
	viewedAt DATETIME NOT NULL,

	PRIMARY KEY(userId,viewName)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
