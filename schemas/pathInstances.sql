/* This table contains a row for each path a user has started. */
CREATE TABLE pathInstances (
	/* Id of this entry. */
	id BIGINT NOT NULL AUTO_INCREMENT,
	/* User who started this path. FK into users. */
	userId VARCHAR(32) NOT NULL,
	/* Id of the page guide that started this path. FK into pageInfos. */
	guideId VARCHAR(32) NOT NULL,
	/* Comma separated list of page ids which this path has. FK into pageInfos. */
	pageIds TEXT NOT NULL,
	/* Comma separated list of which page added the corresponding page to pageIds. FK into pageInfos. */
	sourcePageIds TEXT NOT NULL,
	/* Index of the page the user is on. */
	progress INT NOT NULL,
	/* When this instance was created. */
	createdAt DATETIME NOT NULL,
	/* When this instance was updated last. */
	updatedAt DATETIME NOT NULL,
	/* Optional. If set, the user copied the path from this instance. */
	originalInstanceId BIGINT NOT NULL,
	/* Set to true when the user finished the path. */
	isFinished BOOLEAN NOT NULL,

	PRIMARY KEY(id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
